package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ruziba3vich/mm_api_getway/genprotos/genprotos/article_protos"
	"github.com/ruziba3vich/mm_api_getway/genprotos/genprotos/user_protos"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HTTPHandler struct {
	userClient    user_protos.UserServiceClient
	articleClient article_protos.ArticleServiceClient
}

func NewHTTPHandler(
	userClient user_protos.UserServiceClient,
	articleClient article_protos.ArticleServiceClient,
) *HTTPHandler {
	return &HTTPHandler{
		userClient:    userClient,
		articleClient: articleClient,
	}
}

func (h *HTTPHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		// User routes
		user := api.Group("/users")
		{
			user.POST("/signup", h.SignUp)
			user.POST("/login", h.Login)
			user.GET("/:id", h.GetUserByID)
			user.PUT("/:id", h.UpdateUser)
			user.POST("/:id/follow", h.FollowUser)
			// user.POST("/:id/unfollow", h.UnfollowUser)
			// user.GET("/:id/followers", h.GetUserFollowers)
		}

		// Article routes
		article := api.Group("/articles")
		{
			article.POST("", h.CreateArticle)
			article.GET("", h.GetArticles)
			article.GET("/:id", h.GetArticleByID)
			// article.PUT("/:id", h.UpdateArticle)
			// article.DELETE("/:id", h.DeleteArticle)
			// article.POST("/:id/like", h.LikeArticle)
			// article.POST("/:id/unlike", h.UnlikeArticle)
		}
	}
}

// User Handlers

func (h *HTTPHandler) SignUp(c *gin.Context) {
	var req user_protos.SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.userClient.SignUp(c.Request.Context(), &req)
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *HTTPHandler) Login(c *gin.Context) {
	var req user_protos.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.userClient.Login(c.Request.Context(), &req)
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *HTTPHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.userClient.GetUserData(c.Request.Context(), &user_protos.GetUserDataRequest{
		UserId: id,
	})
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *HTTPHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req user_protos.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.UserId = id

	res, err := h.userClient.UpdateUser(c.Request.Context(), &req)
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *HTTPHandler) FollowUser(c *gin.Context) {
	followerID := c.Param("id")
	var req struct {
		FollowToID string `json:"follow_to_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.userClient.FollowUser(c.Request.Context(), &user_protos.FollowUserRequest{
		FollowerId: followerID,
		FollowToId: req.FollowToID,
	})
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// Article Handlers

func (h *HTTPHandler) CreateArticle(c *gin.Context) {
	var req article_protos.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.articleClient.CreateArticle(c.Request.Context(), &req)
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *HTTPHandler) GetArticles(c *gin.Context) {
	page, limit := getPaginationParams(c)

	res, err := h.articleClient.GetArticles(c.Request.Context(), &article_protos.GetArticlesRequest{
		Pagination: &article_protos.PaginationRequest{
			Page:     page,
			PageSize: limit,
		},
	})
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *HTTPHandler) GetArticleByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.articleClient.GetArticleByID(c.Request.Context(), &article_protos.GetArticleByIDRequest{
		ArticleId: id,
	})
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// Helper functions

func handleGRPCError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	switch st.Code() {
	case codes.NotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
	case codes.InvalidArgument:
		c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
	case codes.AlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": st.Message()})
	case codes.Unauthenticated:
		c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
	}
}

func getPaginationParams(c *gin.Context) (int32, int32) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")

	var pageInt, limitInt int32
	fmt.Sscanf(page, "%d", &pageInt)
	fmt.Sscanf(limit, "%d", &limitInt)

	if pageInt < 1 {
		pageInt = 1
	}
	if limitInt < 1 || limitInt > 100 {
		limitInt = 20
	}

	return pageInt, limitInt
}
