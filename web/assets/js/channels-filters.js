// Filter channels based on current filters
let filteredChannels = []; // 存储筛选后的渠道列表
let modelFilterOptions = [];
let modelFilterCombobox = null; // 通用组件实例

function getModelAllLabel() {
  return (window.t && window.t('channels.modelAll')) || '所有模型';
}

function modelFilterInputValueFromFilterValue(filterValue) {
  if (!filterValue || filterValue === 'all') return getModelAllLabel();
  return filterValue;
}

function normalizeModelFilterOption() {
  if (!filters || !filters.model || filters.model === 'all') return false;
  if (modelFilterOptions.includes(filters.model)) return false;

  filters.model = 'all';
  if (typeof saveChannelsFilters === 'function') saveChannelsFilters();
  return true;
}

function filterChannels() {
  const filtered = channels.filter(channel => {
    if (filters.search && !channel.name.toLowerCase().includes(filters.search.toLowerCase())) {
      return false;
    }

    if (filters.id) {
      const idStr = filters.id.trim();
      if (idStr) {
        const ids = idStr.split(',').map(id => id.trim()).filter(id => id);
        if (ids.length > 0 && !ids.includes(String(channel.id))) {
          return false;
        }
      }
    }

    if (filters.channelType !== 'all') {
      const channelType = channel.channel_type || 'anthropic';
      if (channelType !== filters.channelType) {
        return false;
      }
    }

    if (filters.status !== 'all') {
      if (filters.status === 'enabled' && !channel.enabled) return false;
      if (filters.status === 'disabled' && channel.enabled) return false;
      if (filters.status === 'cooldown' && !(channel.cooldown_remaining_ms > 0)) return false;
    }

    if (filters.model !== 'all') {
      // 新格式：models 是 {model, redirect_model} 对象数组
      const modelNames = Array.isArray(channel.models)
        ? channel.models.map(m => m.model || m)
        : [];
      if (!modelNames.includes(filters.model)) {
        return false;
      }
    }

    return true;
  });

  // 排序：优先使用 effective_priority（健康度模式），否则使用 priority
  filtered.sort((a, b) => {
    const prioA = a.effective_priority ?? a.priority;
    const prioB = b.effective_priority ?? b.priority;
    if (prioB !== prioA) {
      return prioB - prioA;
    }
    const typeA = (a.channel_type || 'anthropic').toLowerCase();
    const typeB = (b.channel_type || 'anthropic').toLowerCase();
    if (typeA !== typeB) {
      return typeA.localeCompare(typeB);
    }
    return a.name.localeCompare(b.name);
  });

  filteredChannels = filtered; // 保存筛选后的列表供其他模块使用

  // 分页切片
  const totalPages = Math.max(1, Math.ceil(filtered.length / channelPageSize));
  if (currentChannelPage > totalPages) currentChannelPage = totalPages;
  if (currentChannelPage < 1) currentChannelPage = 1;
  const startIdx = (currentChannelPage - 1) * channelPageSize;
  const pageChannels = filtered.slice(startIdx, startIdx + channelPageSize);

  renderChannels(pageChannels, startIdx);
  updateChannelPagination(filtered.length);
  updateFilterInfo(filtered.length, channels.length);
}

// Update filter info display
function updateFilterInfo(filtered, total) {
  document.getElementById('filteredCount').textContent = filtered;
  document.getElementById('totalCount').textContent = total;
}

// Update model filter options
function updateModelOptions() {
  const modelSet = new Set();
  const typeFilter = (filters && filters.channelType) ? filters.channelType : 'all';
  channels.forEach(channel => {
    if (typeFilter !== 'all') {
      const channelType = channel.channel_type || 'anthropic';
      if (channelType !== typeFilter) return;
    }
    if (Array.isArray(channel.models)) {
      // 新格式：models 是 {model, redirect_model} 对象数组
      channel.models.forEach(m => {
        const modelName = m.model || m;
        if (modelName) modelSet.add(modelName);
      });
    }
  });

  modelFilterOptions = Array.from(modelSet).sort();

  normalizeModelFilterOption();

  // 使用通用组件刷新下拉框
  if (modelFilterCombobox) {
    modelFilterCombobox.setValue(filters.model, modelFilterInputValueFromFilterValue(filters.model));
    modelFilterCombobox.refresh();
  } else {
    const modelFilterInput = document.getElementById('modelFilter');
    if (modelFilterInput) {
      modelFilterInput.value = modelFilterInputValueFromFilterValue(filters.model);
    }
  }
}

// Setup filter event listeners
function setupFilterListeners() {
  const searchInput = document.getElementById('searchInput');
  const clearSearchBtn = document.getElementById('clearSearchBtn');

  const debouncedFilter = debounce(() => {
    filters.search = searchInput.value;
    if (typeof saveChannelsFilters === 'function') saveChannelsFilters();
    resetChannelPage();
    filterChannels();
    updateClearButton();
  }, 300);

  searchInput.addEventListener('input', debouncedFilter);

  clearSearchBtn.addEventListener('click', () => {
    searchInput.value = '';
    filters.search = '';
    if (typeof saveChannelsFilters === 'function') saveChannelsFilters();
    resetChannelPage();
    filterChannels();
    updateClearButton();
    searchInput.focus();
  });

  function updateClearButton() {
    clearSearchBtn.style.opacity = searchInput.value ? '1' : '0';
  }

  const idFilter = document.getElementById('idFilter');
  const debouncedIdFilter = debounce(() => {
    filters.id = idFilter.value;
    if (typeof saveChannelsFilters === 'function') saveChannelsFilters();
    resetChannelPage();
    filterChannels();
  }, 300);
  idFilter.addEventListener('input', debouncedIdFilter);

  document.getElementById('statusFilter').addEventListener('change', (e) => {
    filters.status = e.target.value;
    if (typeof saveChannelsFilters === 'function') saveChannelsFilters();
    resetChannelPage();
    filterChannels();
  });

  // 使用通用组件初始化模型筛选器（附着模式）
  const modelFilterInput = document.getElementById('modelFilter');
  if (modelFilterInput) {
    modelFilterCombobox = createSearchableCombobox({
      attachMode: true,
      inputId: 'modelFilter',
      dropdownId: 'modelFilterDropdown',
      initialValue: filters.model,
      initialLabel: modelFilterInputValueFromFilterValue(filters.model),
      getOptions: () => {
        const allLabel = getModelAllLabel();
        return [{ value: 'all', label: allLabel }].concat(
          modelFilterOptions.map(m => ({ value: m, label: m }))
        );
      },
      onSelect: (value) => {
        filters.model = value;
        if (typeof saveChannelsFilters === 'function') saveChannelsFilters();
        resetChannelPage();
        filterChannels();
      }
    });
  }

  // 筛选按钮：手动触发筛选
  document.getElementById('btn_filter').addEventListener('click', () => {
    // 收集当前输入框的值
    filters.search = document.getElementById('searchInput').value;
    filters.id = document.getElementById('idFilter').value;

    // 保存筛选条件
    if (typeof saveChannelsFilters === 'function') saveChannelsFilters();

    // 执行筛选
    resetChannelPage();
    filterChannels();
  });

  // 回车键触发筛选
  ['searchInput', 'idFilter'].forEach(id => {
    const el = document.getElementById(id);
    if (el) {
      el.addEventListener('keydown', e => {
        if (e.key === 'Enter') {
          filters.search = document.getElementById('searchInput').value;
          filters.id = document.getElementById('idFilter').value;
          if (typeof saveChannelsFilters === 'function') saveChannelsFilters();
          resetChannelPage();
          filterChannels();
        }
      });
    }
  });
}

// ==================== 分页控制 ====================

function updateChannelPagination(totalFiltered) {
  const paginationEl = document.getElementById('channelPagination');
  if (!paginationEl) return;

  const totalPages = Math.max(1, Math.ceil(totalFiltered / channelPageSize));

  // 仅在有多页时显示分页
  paginationEl.style.display = totalPages > 1 ? '' : 'none';

  const currentPageEl = document.getElementById('ch_current_page');
  const totalPagesEl = document.getElementById('ch_total_pages');
  if (currentPageEl) currentPageEl.textContent = currentChannelPage;
  if (totalPagesEl) totalPagesEl.textContent = totalPages;

  const firstEl = document.getElementById('ch_first');
  const prevEl = document.getElementById('ch_prev');
  const nextEl = document.getElementById('ch_next');
  const lastEl = document.getElementById('ch_last');

  const prevDisabled = currentChannelPage <= 1;
  const nextDisabled = currentChannelPage >= totalPages;

  if (firstEl) firstEl.disabled = prevDisabled;
  if (prevEl) prevEl.disabled = prevDisabled;
  if (nextEl) nextEl.disabled = nextDisabled;
  if (lastEl) lastEl.disabled = nextDisabled;

  // 同步 pageSize 下拉框
  const pageSizeEl = document.getElementById('ch_page_size');
  if (pageSizeEl) pageSizeEl.value = String(channelPageSize);
}

function firstChannelPage() {
  if (currentChannelPage > 1) {
    currentChannelPage = 1;
    filterChannels();
  }
}

function prevChannelPage() {
  if (currentChannelPage > 1) {
    currentChannelPage--;
    filterChannels();
  }
}

function nextChannelPage() {
  const totalPages = Math.max(1, Math.ceil(filteredChannels.length / channelPageSize));
  if (currentChannelPage < totalPages) {
    currentChannelPage++;
    filterChannels();
  }
}

function lastChannelPage() {
  const totalPages = Math.max(1, Math.ceil(filteredChannels.length / channelPageSize));
  if (currentChannelPage < totalPages) {
    currentChannelPage = totalPages;
    filterChannels();
  }
}

function changeChannelPageSize() {
  const el = document.getElementById('ch_page_size');
  if (!el) return;
  const newSize = parseInt(el.value);
  if (newSize > 0 && newSize !== channelPageSize) {
    channelPageSize = newSize;
    currentChannelPage = 1;
    filterChannels();
  }
}

function resetChannelPage() {
  currentChannelPage = 1;
}

function setupChannelPaginationListeners() {
  const firstBtn = document.getElementById('ch_first');
  const prevBtn = document.getElementById('ch_prev');
  const nextBtn = document.getElementById('ch_next');
  const lastBtn = document.getElementById('ch_last');
  const pageSizeSelect = document.getElementById('ch_page_size');

  if (firstBtn) firstBtn.addEventListener('click', firstChannelPage);
  if (prevBtn) prevBtn.addEventListener('click', prevChannelPage);
  if (nextBtn) nextBtn.addEventListener('click', nextChannelPage);
  if (lastBtn) lastBtn.addEventListener('click', lastChannelPage);
  if (pageSizeSelect) pageSizeSelect.addEventListener('change', changeChannelPageSize);
}
