package model

import (
	"errors"
	"strings"
)

// VirtualModel 虚拟模型定义
type VirtualModel struct {
	ID                int64    `json:"id"`
	Name              string   `json:"name"`                       // 虚拟模型名称（唯一）
	Alias             string   `json:"alias,omitempty"`            // 别名
	Description       string   `json:"description,omitempty"`      // 描述
	Enabled           bool     `json:"enabled"`                    // 是否启用
	DefaultFallback   string   `json:"default_fallback,omitempty"` // 默认回退模型
	AssociationsCount int      `json:"associations_count"`         // 关联规则数量
	CreatedAt         JSONTime `json:"created_at"`
	UpdatedAt         JSONTime `json:"updated_at"`
}

// Validate 验证虚拟模型字段
func (v *VirtualModel) Validate() error {
	v.Name = strings.TrimSpace(v.Name)
	if v.Name == "" {
		return errors.New("virtual model name cannot be empty")
	}
	if len(v.Name) > 128 {
		return errors.New("virtual model name too long (max 128 characters)")
	}
	if strings.ContainsAny(v.Name, "\x00\r\n") {
		return errors.New("virtual model name contains illegal characters")
	}

	v.Alias = strings.TrimSpace(v.Alias)
	if len(v.Alias) > 128 {
		return errors.New("alias too long (max 128 characters)")
	}

	v.DefaultFallback = strings.TrimSpace(v.DefaultFallback)
	if len(v.DefaultFallback) > 256 {
		return errors.New("default fallback too long (max 256 characters)")
	}

	return nil
}

// MatchType 匹配类型
type MatchType string

const (
	MatchTypeExact    MatchType = "exact"    // 精确匹配
	MatchTypePrefix   MatchType = "prefix"   // 前缀匹配
	MatchTypeSuffix   MatchType = "suffix"   // 后缀匹配
	MatchTypeContains MatchType = "contains" // 包含匹配
	MatchTypeRegex    MatchType = "regex"    // 正则匹配
	MatchTypeWildcard MatchType = "wildcard" // 通配符匹配
)

// IsValid 检查匹配类型是否有效
func (m MatchType) IsValid() bool {
	switch m {
	case MatchTypeExact, MatchTypePrefix, MatchTypeSuffix, MatchTypeContains, MatchTypeRegex, MatchTypeWildcard:
		return true
	default:
		return false
	}
}

// ModelAssociation 模型关联规则
type ModelAssociation struct {
	ID             int64     `json:"id"`
	VirtualModelID int64     `json:"virtual_model_id"` // 虚拟模型ID
	ChannelID      int64     `json:"channel_id"`       // 渠道ID
	MatchType      MatchType `json:"match_type"`       // 匹配类型
	Pattern        string    `json:"pattern"`          // 匹配模式
	Priority       int       `json:"priority"`         // 优先级（数值越大优先级越高）
	Enabled        bool      `json:"enabled"`          // 是否启用
	CreatedAt      JSONTime  `json:"created_at"`
	UpdatedAt      JSONTime  `json:"updated_at"`
}

// Validate 验证模型关联规则
func (m *ModelAssociation) Validate() error {
	if m.VirtualModelID <= 0 {
		return errors.New("virtual_model_id must be positive")
	}
	if m.ChannelID <= 0 {
		return errors.New("channel_id must be positive")
	}
	if !m.MatchType.IsValid() {
		return errors.New("invalid match_type")
	}

	m.Pattern = strings.TrimSpace(m.Pattern)
	if m.Pattern == "" {
		return errors.New("pattern cannot be empty")
	}
	if len(m.Pattern) > 256 {
		return errors.New("pattern too long (max 256 characters)")
	}
	if strings.ContainsAny(m.Pattern, "\x00\r\n") {
		return errors.New("pattern contains illegal characters")
	}

	return nil
}

// ModelAssociationWithDetails 带详情的模型关联规则（用于API响应）
type ModelAssociationWithDetails struct {
	ModelAssociation
	VirtualModelName string `json:"virtual_model_name,omitempty"` // 虚拟模型名称
	ChannelName      string `json:"channel_name,omitempty"`       // 渠道名称
}
