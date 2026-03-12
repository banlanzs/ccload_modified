package app

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"ccLoad+ccr/internal/model"
	"ccLoad+ccr/internal/storage"
)

// HealthCache 渠道健康度缓存
type HealthCache struct {
	store  storage.Store
	config model.HealthScoreConfig
	mu     sync.RWMutex // 保护 config 字段的并发访问

	// 健康统计缓存：使用原子指针实现无锁快照替换
	// 读取时直接Load，更新时用新map整体替换，避免遍历删除的并发问题
	healthStats atomic.Pointer[map[int64]model.ChannelHealthStats]

	// 控制
	stopCh   chan struct{}
	innerCh  chan struct{} // 内部停止信号，用于启停切换
	wg       *sync.WaitGroup

	// shutdown标志
	isShuttingDown *atomic.Bool

	// 运行状态（用于幂等性检查）
	running atomic.Bool
}

// NewHealthCache 创建健康度缓存
func NewHealthCache(store storage.Store, config model.HealthScoreConfig, shutdownCh chan struct{}, isShuttingDown *atomic.Bool, wg *sync.WaitGroup) *HealthCache {
	h := &HealthCache{
		store:          store,
		config:         config,
		stopCh:         shutdownCh,
		innerCh:        make(chan struct{}),
		wg:             wg,
		isShuttingDown: isShuttingDown,
	}
	// 初始化空map
	emptyMap := make(map[int64]model.ChannelHealthStats)
	h.healthStats.Store(&emptyMap)
	return h
}

// Start 启动后台更新协程
func (h *HealthCache) Start() {
	h.mu.RLock()
	enabled := h.config.Enabled
	updateInterval := h.config.UpdateIntervalSeconds
	windowMinutes := h.config.WindowMinutes
	h.mu.RUnlock()

	if !enabled {
		return
	}
	if updateInterval <= 0 || windowMinutes <= 0 {
		log.Printf("[WARN] 健康度缓存未启动：无效配置 update_interval=%d window_minutes=%d", updateInterval, windowMinutes)
		return
	}

	// 幂等性检查：已启动则跳过
	if h.running.Swap(true) {
		log.Print("[INFO] 健康度缓存更新循环已在运行，跳过重复启动")
		return
	}

	h.wg.Add(1)
	go h.updateLoop()
	log.Print("[INFO] 健康度缓存更新循环已启动")
}

// updateLoop 定期更新成功率缓存
func (h *HealthCache) updateLoop() {
	defer h.wg.Done()
	defer h.running.Store(false)

	// 立即执行一次
	h.update()

	h.mu.RLock()
	interval := time.Duration(h.config.UpdateIntervalSeconds) * time.Second
	h.mu.RUnlock()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-h.stopCh:
			return
		case <-h.innerCh:
			log.Print("[INFO] 健康度缓存更新循环收到停止信号")
			return
		case <-ticker.C:
			if h.isShuttingDown.Load() {
				return
			}
			h.update()
		}
	}
}

// update 更新成功率缓存
func (h *HealthCache) update() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	h.mu.RLock()
	windowMinutes := h.config.WindowMinutes
	h.mu.RUnlock()

	since := time.Now().Add(-time.Duration(windowMinutes) * time.Minute)
	stats, err := h.store.GetChannelSuccessRates(ctx, since)
	if err != nil {
		log.Printf("[WARN] 更新渠道成功率缓存失败: %v", err)
		return
	}

	// 原子替换：用新快照整体替换旧数据，避免遍历删除的并发问题
	h.healthStats.Store(&stats)
}

// GetHealthStats 获取渠道健康统计，不存在返回默认值（新渠道不惩罚）
func (h *HealthCache) GetHealthStats(channelID int64) model.ChannelHealthStats {
	stats := h.healthStats.Load()
	if stats == nil {
		return model.ChannelHealthStats{SuccessRate: 1.0, SampleCount: 0}
	}
	if v, ok := (*stats)[channelID]; ok {
		return v
	}
	return model.ChannelHealthStats{SuccessRate: 1.0, SampleCount: 0} // 新渠道默认成功率100%
}

// GetSuccessRate 获取渠道成功率（兼容旧接口）
func (h *HealthCache) GetSuccessRate(channelID int64) float64 {
	return h.GetHealthStats(channelID).SuccessRate
}

// GetAllSuccessRates 获取所有渠道成功率（返回快照副本，兼容旧接口）
func (h *HealthCache) GetAllSuccessRates() map[int64]float64 {
	stats := h.healthStats.Load()
	if stats == nil {
		return make(map[int64]float64)
	}
	result := make(map[int64]float64, len(*stats))
	for k, v := range *stats {
		result[k] = v.SuccessRate
	}
	return result
}

// Config 返回健康度配置
func (h *HealthCache) Config() model.HealthScoreConfig {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.config
}

// UpdateConfig 动态更新健康度配置（支持热更新）
func (h *HealthCache) UpdateConfig(newConfig model.HealthScoreConfig) error {
	h.mu.Lock()
	oldEnabled := h.config.Enabled
	h.config = newConfig
	h.mu.Unlock()

	// 处理启停切换
	if !oldEnabled && newConfig.Enabled {
		// 从禁用变为启用：启动更新循环
		log.Print("[INFO] 健康度排序已启用，启动更新循环")
		h.Start()
	} else if oldEnabled && !newConfig.Enabled {
		// 从启用变为禁用：停止更新循环
		log.Print("[INFO] 健康度排序已禁用，停止更新循环")
		h.stop()
	} else if oldEnabled && newConfig.Enabled {
		// 两者都启用：配置参数可能变化，记录日志
		log.Print("[INFO] 健康度配置已更新（保持启用状态）")
	}

	return nil
}

// stop 停止后台更新循环（内部方法）
func (h *HealthCache) stop() {
	if !h.running.Load() {
		return
	}
	// 重置 innerCh 以发送停止信号
	close(h.innerCh)
	h.innerCh = make(chan struct{})
}
