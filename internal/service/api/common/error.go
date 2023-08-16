package common

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser/pcerr"
)

func AbortWithError(c *gin.Context, err error) {
	// As per RFC7807 problem details should set the content type as application/problem+json
	// Openapi does not allow to specify different content type based on the response.
	// For now we will use the application/json instead of application/problem+json
	// c.Header("Content-Type", "application/problem+json")
	c.Header("Content-Type", "application/json")

	var requestURL *url.URL

	if c.Request != nil {
		requestURL = c.Request.URL
	}

	switch e := err.(type) {
	case pcerr.ValidationError:
		badRequest := pcerr.BadRequestWithTypeAndMessage(pcerr.ErrorCodeInvalidData, "Validation Errors")
		badRequest.ValidationErrors = []*pcerr.ErrorDetailModel{}
		for _, v := range e.Errors {
			badRequest.ValidationErrors = append(badRequest.ValidationErrors, &pcerr.ErrorDetailModel{Message: detailMessageByTag(v), Location: fmt.Sprintf("%s.%s", e.Type, getNormalizedField(v.Namespace()))})
		}
		fplog.Logger(c).Error("Validation error",
			"error", badRequest,
			"errorID", badRequest.Instance,
			"detail", badRequest.ValidationErrors,
			"requestURL", requestURL)
		c.AbortWithStatusJSON(http.StatusBadRequest, badRequest)
	case validator.ValidationErrors:
		badRequest := pcerr.BadRequestWithTypeAndMessage(pcerr.ErrorCodeInvalidData, "Validation Errors")
		badRequest.ValidationErrors = []*pcerr.ErrorDetailModel{}
		for _, v := range e {
			badRequest.ValidationErrors = append(badRequest.ValidationErrors, &pcerr.ErrorDetailModel{Message: detailMessageByTag(v), Location: getNormalizedField(v.Namespace())})
		}
		fplog.Logger(c).Error("Validation error",
			"error", badRequest,
			"errorID", badRequest.Instance,
			"detail", badRequest.ValidationErrors,
			"requestURL", requestURL)
		c.AbortWithStatusJSON(http.StatusBadRequest, badRequest)
	case pcerr.ErrorModel:
		fplog.Logger(c).Error("Error "+e.Instance,
			"error", e,
			"errorID", e.Instance,
			"requestURL", requestURL)
		c.AbortWithStatusJSON(e.Status, e)
	default:
		newErr := pcerr.InternalWithMessage("Internal Error.")
		fplog.Logger(c).Error("Error "+newErr.Instance,
			"error", newErr,
			"errorID", newErr.Instance,
			"originalError", err,
			"requestURL", requestURL)
		c.AbortWithStatusJSON(http.StatusInternalServerError, newErr)
	}
}

func getNormalizedField(namespace string) string {
	elements := strings.Split(namespace, ".")
	var index int
	for i, element := range elements {
		if strings.ToLower(element) == element {
			index = i
			break
		}
	}
	return strings.Join(elements[index:], ".")
}

func detailMessageByTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "min":
		if fe.Param() == "1" {
			return fmt.Sprintf("%s cannot be empty.", fe.Field())
		}
		return fmt.Sprintf("%s must have a minimum length of %s.", fe.Field(), fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of %s.", fe.Field(), prettifyOneOfParams(fe.Param()))
	case "required":
		return fmt.Sprintf("%s is required.", fe.Field())
	case "steampipe_tags":
		return fmt.Sprintf("%s is not a valid tags format.", fe.Field())
	case "steampipe_workspace_api_handle":
		return fmt.Sprintf("%s is invalid.", fe.Field())
	case "steampipe_identity_token_min_issued_at":
		return "token_min_issued_at must have a value less than or equal to the current time."
	}
	return fe.Error()

}

func prettifyOneOfParams(input string) string {
	var items []string
	for _, item := range strings.Split(input, " ") {
		items = append(items, fmt.Sprintf("'%s'", item))
	}
	return strings.Join(items, ", ")
}
