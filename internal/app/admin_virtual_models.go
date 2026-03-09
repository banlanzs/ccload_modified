package app

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"ccLoad+ccr/internal/model"

	"github.com/gin-gonic/gin"
)

// ============================================================================
// 虚拟模型管理 (Admin API)
// ============================================================================

// HandleListVirtualModels 列出所有虚拟模型
// GET /admin/virtual-models
func (s *Server) HandleListVirtualModels(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	models, err := s.store.ListVirtualModels(ctx)
	if err != nil {
		log.Print("[ERROR] 列出虚拟模型失败: " + err.Error())
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	if models == nil {
		models = make([]*model.VirtualModel, 0)
	}

	RespondJSON(c, http.StatusOK, models)
}

// HandleGetVirtualModel 获取虚拟模型详情
// GET /admin/virtual-models/:id
func (s *Server) HandleGetVirtualModel(c *gin.Context) {
	id, err := ParseInt64Param(c, "id")
	if err != nil {
		RespondErrorMsg(c, http.StatusBadRequest, "invalid virtual model id")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	vm, err := s.store.GetVirtualModel(ctx, id)
	if err != nil {
		RespondErrorMsg(c, http.StatusNotFound, "virtual model not found")
		return
	}

	RespondJSON(c, http.StatusOK, vm)
}

// HandleCreateVirtualModel 创建虚拟模型
// POST /admin/virtual-models
func (s *Server) HandleCreateVirtualModel(c *gin.Context) {
	var req model.VirtualModel
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		RespondErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	created, err := s.store.CreateVirtualModel(ctx, &req)
	if err != nil {
		log.Print("[ERROR] 创建虚拟模型失败: " + err.Error())
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	log.Printf("[INFO] 创建虚拟模型: ID=%d, 名称=%s", created.ID, created.Name)
	RespondJSON(c, http.StatusOK, created)
}

// HandleUpdateVirtualModel 更新虚拟模型
// PUT /admin/virtual-models/:id
func (s *Server) HandleUpdateVirtualModel(c *gin.Context) {
	id, err := ParseInt64Param(c, "id")
	if err != nil {
		RespondErrorMsg(c, http.StatusBadRequest, "invalid virtual model id")
		return
	}

	var req model.VirtualModel
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		RespondErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := s.store.UpdateVirtualModel(ctx, id, &req); err != nil {
		log.Print("[ERROR] 更新虚拟模型失败: " + err.Error())
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	// 获取更新后的模型
	updated, err := s.store.GetVirtualModel(ctx, id)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	log.Printf("[INFO] 更新虚拟模型: ID=%d, 名称=%s", id, updated.Name)
	RespondJSON(c, http.StatusOK, updated)
}

// HandleDeleteVirtualModel 删除虚拟模型
// DELETE /admin/virtual-models/:id
func (s *Server) HandleDeleteVirtualModel(c *gin.Context) {
	id, err := ParseInt64Param(c, "id")
	if err != nil {
		RespondErrorMsg(c, http.StatusBadRequest, "invalid virtual model id")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := s.store.DeleteVirtualModel(ctx, id); err != nil {
		log.Print("[ERROR] 删除虚拟模型失败: " + err.Error())
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	log.Printf("[INFO] 删除虚拟模型: ID=%d", id)
	RespondJSON(c, http.StatusOK, gin.H{"id": id})
}

// ============================================================================
// 模型关联规则管理 (Admin API)
// ============================================================================

// HandleListModelAssociations 列出模型关联规则
// GET /admin/model-associations?virtual_model_id=1
func (s *Server) HandleListModelAssociations(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	virtualModelIDStr := c.Query("virtual_model_id")
	if virtualModelIDStr != "" {
		virtualModelID, err := strconv.ParseInt(virtualModelIDStr, 10, 64)
		if err != nil {
			RespondErrorMsg(c, http.StatusBadRequest, "invalid virtual_model_id")
			return
		}

		associations, err := s.store.ListModelAssociationsWithDetails(ctx, virtualModelID)
		if err != nil {
			log.Print("[ERROR] 列出模型关联规则失败: " + err.Error())
			RespondError(c, http.StatusInternalServerError, err)
			return
		}

		if associations == nil {
			associations = make([]*model.ModelAssociationWithDetails, 0)
		}

		RespondJSON(c, http.StatusOK, associations)
		return
	}

	// 列出所有关联规则
	associations, err := s.store.ListAllModelAssociations(ctx)
	if err != nil {
		log.Print("[ERROR] 列出所有模型关联规则失败: " + err.Error())
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	if associations == nil {
		associations = make([]*model.ModelAssociation, 0)
	}

	RespondJSON(c, http.StatusOK, associations)
}

// HandleGetModelAssociation 获取模型关联规则详情
// GET /admin/model-associations/:id
func (s *Server) HandleGetModelAssociation(c *gin.Context) {
	id, err := ParseInt64Param(c, "id")
	if err != nil {
		RespondErrorMsg(c, http.StatusBadRequest, "invalid association id")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	ma, err := s.store.GetModelAssociation(ctx, id)
	if err != nil {
		RespondErrorMsg(c, http.StatusNotFound, "model association not found")
		return
	}

	RespondJSON(c, http.StatusOK, ma)
}

// HandleCreateModelAssociation 创建模型关联规则
// POST /admin/model-associations
func (s *Server) HandleCreateModelAssociation(c *gin.Context) {
	var req model.ModelAssociation
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		RespondErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	created, err := s.store.CreateModelAssociation(ctx, &req)
	if err != nil {
		log.Print("[ERROR] 创建模型关联规则失败: " + err.Error())
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	log.Printf("[INFO] 创建模型关联规则: ID=%d, 虚拟模型ID=%d, 渠道ID=%d", created.ID, created.VirtualModelID, created.ChannelID)
	RespondJSON(c, http.StatusOK, created)
}

// HandleUpdateModelAssociation 更新模型关联规则
// PUT /admin/model-associations/:id
func (s *Server) HandleUpdateModelAssociation(c *gin.Context) {
	id, err := ParseInt64Param(c, "id")
	if err != nil {
		RespondErrorMsg(c, http.StatusBadRequest, "invalid association id")
		return
	}

	var req model.ModelAssociation
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		RespondErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := s.store.UpdateModelAssociation(ctx, id, &req); err != nil {
		log.Print("[ERROR] 更新模型关联规则失败: " + err.Error())
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	// 获取更新后的规则
	updated, err := s.store.GetModelAssociation(ctx, id)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	log.Printf("[INFO] 更新模型关联规则: ID=%d", id)
	RespondJSON(c, http.StatusOK, updated)
}

// HandleDeleteModelAssociation 删除模型关联规则
// DELETE /admin/model-associations/:id
func (s *Server) HandleDeleteModelAssociation(c *gin.Context) {
	id, err := ParseInt64Param(c, "id")
	if err != nil {
		RespondErrorMsg(c, http.StatusBadRequest, "invalid association id")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := s.store.DeleteModelAssociation(ctx, id); err != nil {
		log.Print("[ERROR] 删除模型关联规则失败: " + err.Error())
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	log.Printf("[INFO] 删除模型关联规则: ID=%d", id)
	RespondJSON(c, http.StatusOK, gin.H{"id": id})
}

// HandleValidateModelAssociations 校验模型关联规则（冲突检测）
// POST /admin/model-associations/validate
func (s *Server) HandleValidateModelAssociations(c *gin.Context) {
	var req struct {
		VirtualModelID int64             `json:"virtual_model_id" binding:"required"`
		ChannelID      int64             `json:"channel_id" binding:"required"`
		MatchType      model.MatchType   `json:"match_type" binding:"required"`
		Pattern        string            `json:"pattern" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		RespondErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	// 基础验证
	if !req.MatchType.IsValid() {
		RespondErrorMsg(c, http.StatusBadRequest, "invalid match_type")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 获取同一虚拟模型的所有规则
	associations, err := s.store.ListModelAssociations(ctx, req.VirtualModelID)
	if err != nil {
		log.Print("[ERROR] 查询模型关联规则失败: " + err.Error())
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	// 检查冲突
	conflicts := make([]gin.H, 0)
	for _, ma := range associations {
		if ma.ChannelID == req.ChannelID && ma.MatchType == req.MatchType && ma.Pattern == req.Pattern {
			conflicts = append(conflicts, gin.H{
				"id":         ma.ID,
				"channel_id": ma.ChannelID,
				"match_type": ma.MatchType,
				"pattern":    ma.Pattern,
				"priority":   ma.Priority,
			})
		}
	}

	RespondJSON(c, http.StatusOK, gin.H{
		"valid":     len(conflicts) == 0,
		"conflicts": conflicts,
	})
}

// HandlePreviewModelAssociations 预览路由结果
// POST /admin/model-associations/preview
func (s *Server) HandlePreviewModelAssociations(c *gin.Context) {
	var req struct {
		Model       string `json:"model" binding:"required"`
		RequestType string `json:"request_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		RespondErrorMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 使用 ModelResolver 解析虚拟模型并获取候选
	candidates, err := s.modelResolver.Resolve(ctx, req.Model, req.RequestType)
	if err != nil {
		log.Print("[ERROR] 预览路由解析失败: " + err.Error())
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	// 转换为响应格式
	matchedRules := make([]gin.H, 0)
	resultCandidates := make([]gin.H, 0)

	for _, cand := range candidates {
		matchedRules = append(matchedRules, gin.H{
			"rule_id":      cand.RuleID,
			"channel_id":   cand.ChannelID,
			"channel_name": cand.ChannelName,
			"match_type":   cand.MatchType,
			"pattern":      cand.Pattern,
			"priority":     cand.Priority,
		})

		resultCandidates = append(resultCandidates, gin.H{
			"channel_id":     cand.ChannelID,
			"channel_name":   cand.ChannelName,
			"virtual_model": cand.VirtualModel,
			"resolved_model": cand.ResolvedModel,
		})
	}

	RespondJSON(c, http.StatusOK, gin.H{
		"model":          req.Model,
		"request_type":   req.RequestType,
		"matched_rules":  matchedRules,
		"candidates":     resultCandidates,
		"message":        "Preview功能已启用",
	})
}
