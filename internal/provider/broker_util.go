package provider

import (
	"fmt"
	"regexp"

	"github.com/clbanning/mxj/v2"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

/** helper for handling defaults, returns nil instead of ponter to "" for empty strings */
func nullIfEmptyStringPtr(s basetypes.StringValue) *string {
	if s.ValueString() != "" {
		return s.ValueStringPointer()
	}
	return nil
}

/** helper for handling defaults, returns nil for inknwon int32 */
func nullIfEmptyInt32Ptr(v basetypes.Int32Value) *int32 {
	if v.IsUnknown() {
		return nil
	}
	return v.ValueInt32Pointer()
}

// extract error infos from ErrorDTO
func parseErrorDTO(body []byte) string {
	m, err := mxj.NewMapXml(body)
	if err != nil {
		// just return the full response
		return string(body)
	}
	return fmt.Sprintf("Message: %s\nValidationDetails: %s\n",
		m["ErrorDTO"].(map[string]interface{})["message"],
		m["ErrorDTO"].(map[string]interface{})["validationDetails"])

}

// helper to extract the router prefix from the router name
func getRouterPrefix(routerName string) string {
	re := regexp.MustCompile(`^(.*)(primary|backup|monitoring)+(cn)?`)
	return re.ReplaceAllString(routerName, "$1")
}
