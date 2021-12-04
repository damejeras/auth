package application

import "net/http"

type Middleware func(handlerFunc http.HandlerFunc) http.HandlerFunc
