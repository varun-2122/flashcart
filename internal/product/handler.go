package product

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/domain"
	"github.com/varun-2122/flashcart/internal/response"
)

type ProductHandler struct {
	service *ProductService
}

func NewProductHandler(service *ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", err.Error())
		return
	}

	p, err := h.service.CreateProduct(r.Context(), req)
	if err != nil {
		response.BadRequest(w, err.Error(), nil)
		return
	}

	response.Created(w, p)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid product UUID", nil)
		return
	}

	p, err := h.service.GetProductByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			response.NotFound(w, "Product not found")
			return
		}
		response.InternalServerError(w, err.Error())
		return
	}

	response.Success(w, p)
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := domain.ProductFilter{
		Brand:  q.Get("brand"),
		Search: q.Get("search"),
		Limit:  limit,
		Offset: offset,
	}

	if catStr := q.Get("category_id"); catStr != "" {
		if catID, err := uuid.Parse(catStr); err == nil {
			filter.CategoryID = &catID
		}
	}

	if minStr := q.Get("min_price"); minStr != "" {
		if minP, err := strconv.ParseFloat(minStr, 64); err == nil {
			filter.MinPrice = &minP
		}
	}

	if maxStr := q.Get("max_price"); maxStr != "" {
		if maxP, err := strconv.ParseFloat(maxStr, 64); err == nil {
			filter.MaxPrice = &maxP
		}
	}

	products, total, err := h.service.ListProducts(r.Context(), filter)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"products": products,
		"meta": map[string]any{
			"total":  total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		},
	})
}
