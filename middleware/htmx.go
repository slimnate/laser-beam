package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type HTMXHeaders struct {
	Boost                 bool
	Request               bool
	HistoryRestoreRequest bool

	CurrentUrl  string
	Prompt      string
	Target      string
	TriggerName string
	Trigger     string
}

func HTMXMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		hxBoostHeader := ctx.GetHeader("HX-Boost")
		hxRequestHeader := ctx.GetHeader("HX-Request")
		hxHistoryRestoreRequestHeader := ctx.GetHeader("HX-History-Restore-Request")
		hxCurrentUrlHeader := ctx.GetHeader("HX-Current-Url")
		hxPromptHeader := ctx.GetHeader("HX-Prompt")
		hxTargetHeader := ctx.GetHeader("HX-Target")
		hxTriggerHeader := ctx.GetHeader("HX-Trigger")
		hxTriggerNameHeader := ctx.GetHeader("HX-Trigger-Name")

		hxBoost, _ := strconv.ParseBool(hxBoostHeader)
		hxRequest, _ := strconv.ParseBool(hxRequestHeader)
		hxHistoryRestoreRequest, _ := strconv.ParseBool(hxHistoryRestoreRequestHeader)

		ctx.Set("hx", &HTMXHeaders{
			Boost:                 hxBoost,
			Request:               hxRequest,
			HistoryRestoreRequest: hxHistoryRestoreRequest,
			CurrentUrl:            hxCurrentUrlHeader,
			Prompt:                hxPromptHeader,
			Target:                hxTargetHeader,
			Trigger:               hxTriggerHeader,
			TriggerName:           hxTriggerNameHeader,
		})

		ctx.Next()
	}
}

func GetHxHeaders(ctx *gin.Context) *HTMXHeaders {
	hxHeaders, _ := ctx.Get("hx")
	return hxHeaders.(*HTMXHeaders)
}
