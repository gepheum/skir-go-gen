package skir_client

// Method identifies one method in a Skir service, both on the server side and
// the client side.
//
//   - Request is the type of the request parameter.
//   - Response is the type of the response returned by this method.
type Method[Request, Response any] struct {
	// Name is the name of the method.
	Name string
	// Number is the unique number identifying this method within the service.
	Number int64
	// RequestSerializer serializes and deserializes the request type.
	RequestSerializer Serializer[Request]
	// ResponseSerializer serializes and deserializes the response type.
	ResponseSerializer Serializer[Response]
	// Doc is the documentation string for this method.
	Doc string
}
