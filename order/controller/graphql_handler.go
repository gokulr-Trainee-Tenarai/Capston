package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"order/model"
	"order/service"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
)

type GraphQLRequestBody struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

type GraphQLHandler struct {
	schema graphql.Schema
}

func NewGraphQLHandler(svc service.OrderService) (*GraphQLHandler, error) {
	orderType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Order",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if order, ok := p.Source.(model.Order); ok {
						return int(order.ID), nil
					}
					if orderPtr, ok := p.Source.(*model.Order); ok && orderPtr != nil {
						return int(orderPtr.ID), nil
					}
					return nil, nil
				},
			},
			"customerId": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if order, ok := p.Source.(model.Order); ok {
						return int(order.CustomerID), nil
					}
					if orderPtr, ok := p.Source.(*model.Order); ok && orderPtr != nil {
						return int(orderPtr.CustomerID), nil
					}
					return nil, nil
				},
			},
			"productId": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if order, ok := p.Source.(model.Order); ok {
						return int(order.ProductID), nil
					}
					if orderPtr, ok := p.Source.(*model.Order); ok && orderPtr != nil {
						return int(orderPtr.ProductID), nil
					}
					return nil, nil
				},
			},
			"quantity": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if order, ok := p.Source.(model.Order); ok {
						return order.Quantity, nil
					}
					if orderPtr, ok := p.Source.(*model.Order); ok && orderPtr != nil {
						return orderPtr.Quantity, nil
					}
					return nil, nil
				},
			},
			"unitPrice": &graphql.Field{
				Type: graphql.Float,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if order, ok := p.Source.(model.Order); ok {
						return order.UnitPrice, nil
					}
					if orderPtr, ok := p.Source.(*model.Order); ok && orderPtr != nil {
						return orderPtr.UnitPrice, nil
					}
					return nil, nil
				},
			},
			"totalAmount": &graphql.Field{
				Type: graphql.Float,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if order, ok := p.Source.(model.Order); ok {
						return order.TotalAmount, nil
					}
					if orderPtr, ok := p.Source.(*model.Order); ok && orderPtr != nil {
						return orderPtr.TotalAmount, nil
					}
					return nil, nil
				},
			},
			"status": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if order, ok := p.Source.(model.Order); ok {
						return string(order.Status), nil
					}
					if orderPtr, ok := p.Source.(*model.Order); ok && orderPtr != nil {
						return string(orderPtr.Status), nil
					}
					return nil, nil
				},
			},
			"createdAt": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if order, ok := p.Source.(model.Order); ok {
						return order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), nil
					}
					if orderPtr, ok := p.Source.(*model.Order); ok && orderPtr != nil {
						return orderPtr.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), nil
					}
					return nil, nil
				},
			},
			"updatedAt": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if order, ok := p.Source.(model.Order); ok {
						return order.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"), nil
					}
					if orderPtr, ok := p.Source.(*model.Order); ok && orderPtr != nil {
						return orderPtr.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"), nil
					}
					return nil, nil
				},
			},
		},
	})

	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"getOrder": &graphql.Field{
				Type:        orderType,
				Description: "Get order by ID",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id := p.Args["id"].(int)
					return svc.GetOrderByID(p.Context, uint(id))
				},
			},
			"listOrders": &graphql.Field{
				Type:        graphql.NewList(orderType),
				Description: "List all orders",
				Args: graphql.FieldConfigArgument{
					"limit":  &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 10},
					"offset": &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 0},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					limit := p.Args["limit"].(int)
					offset := p.Args["offset"].(int)
					return svc.ListOrders(p.Context, limit, offset)
				},
			},
		},
	})

	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createOrder": &graphql.Field{
				Type:        orderType,
				Description: "Create a new order",
				Args: graphql.FieldConfigArgument{
					"customerId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
					"productId":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
					"quantity":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					input := model.CreateOrderInput{
						CustomerID: uint(p.Args["customerId"].(int)),
						ProductID:  uint(p.Args["productId"].(int)),
						Quantity:   p.Args["quantity"].(int),
					}
					return svc.CreateOrder(p.Context, input)
				},
			},
			"updateOrderStatus": &graphql.Field{
				Type:        orderType,
				Description: "Update order status",
				Args: graphql.FieldConfigArgument{
					"id":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
					"status": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id := uint(p.Args["id"].(int))
					status := model.OrderStatus(p.Args["status"].(string))
					if err := svc.UpdateOrderStatus(p.Context, id, status); err != nil {
						return nil, err
					}
					return svc.GetOrderByID(p.Context, id)
				},
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
	if err != nil {
		return nil, err
	}

	return &GraphQLHandler{schema: schema}, nil
}

func (h *GraphQLHandler) Handle(c *gin.Context) {
	var req GraphQLRequestBody

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil || len(bodyBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"errors": []string{"empty request body"}})
		return
	}

	if jsonErr := json.Unmarshal(bodyBytes, &req); jsonErr != nil || strings.TrimSpace(req.Query) == "" {
		rawQuery := strings.TrimSpace(string(bodyBytes))
		rawQuery = strings.TrimPrefix(rawQuery, "\"")
		rawQuery = strings.TrimSuffix(rawQuery, "\"")
		req.Query = rawQuery
	}

	if strings.TrimSpace(req.Query) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"errors": []string{"missing GraphQL operation"}})
		return
	}

	result := graphql.Do(graphql.Params{
		Schema:         h.schema,
		RequestString:  req.Query,
		VariableValues: req.Variables,
		OperationName:  req.OperationName,
		Context:        c.Request.Context(),
	})

	if len(result.Errors) > 0 {
		c.JSON(http.StatusBadRequest, result)
		return
	}

	c.JSON(http.StatusOK, result)
}
