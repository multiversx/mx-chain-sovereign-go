package groups

import (
	"strings"

	"github.com/gin-gonic/gin"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/config"
)

var log = logger.GetOrCreate("api/groups")

type endpointProperties struct {
	isOpen bool
}

// BaseGroup defines the base group
type BaseGroup struct {
	Endpoints []*shared.EndpointHandlerData
}

// GetEndpoints returns all the Endpoints specific to the group
func (bg *BaseGroup) GetEndpoints() []*shared.EndpointHandlerData {
	return bg.Endpoints
}

// RegisterRoutes will register all the Endpoints to the given web server
func (bg *BaseGroup) RegisterRoutes(
	ws *gin.RouterGroup,
	apiConfig config.ApiRoutesConfig,
) {
	for _, handlerData := range bg.Endpoints {
		properties := getEndpointProperties(ws, handlerData.Path, apiConfig)

		if !properties.isOpen {
			log.Debug("endpoint is closed", "path", handlerData.Path)
			continue
		}

		middlewares := make([]gin.HandlerFunc, 0)
		beforeSpecifiesMiddlewares, afterSpecificMiddlewares := extractSpecificMiddlewares(handlerData.AdditionalMiddlewares)

		middlewares = append(middlewares, beforeSpecifiesMiddlewares...)
		middlewares = append(middlewares, handlerData.Handler)
		middlewares = append(middlewares, afterSpecificMiddlewares...)

		ws.Handle(handlerData.Method, handlerData.Path, middlewares...)
	}
}

func extractSpecificMiddlewares(middlewares []shared.AdditionalMiddleware) ([]gin.HandlerFunc, []gin.HandlerFunc) {
	if len(middlewares) == 0 {
		return nil, nil
	}

	beforeMiddlewares := make([]gin.HandlerFunc, 0)
	afterMiddlewares := make([]gin.HandlerFunc, 0)
	for _, middleware := range middlewares {
		if middleware.Position == shared.After {
			afterMiddlewares = append(afterMiddlewares, middleware.Middleware)
		} else {
			beforeMiddlewares = append(beforeMiddlewares, middleware.Middleware)
		}
	}

	return beforeMiddlewares, afterMiddlewares
}

func getEndpointProperties(ws *gin.RouterGroup, path string, apiConfig config.ApiRoutesConfig) endpointProperties {
	basePath := ws.BasePath()

	// ws.BasePath will return paths like /group or /v1.0/group, so we need the last token after splitting by /
	splitPath := strings.Split(basePath, "/")
	basePath = splitPath[len(splitPath)-1]

	group, ok := apiConfig.APIPackages[basePath]
	if !ok {
		return endpointProperties{
			isOpen: false,
		}
	}

	for _, route := range group.Routes {
		if route.Name == path {
			return endpointProperties{
				isOpen: route.Open,
			}
		}
	}

	return endpointProperties{
		isOpen: false,
	}
}
