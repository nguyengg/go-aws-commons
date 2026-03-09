package config

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Config contains options that can be provided either at construction, or overridden when calling methods that make use
// of Config.
type Config struct {
	// Client is the client for making DynamoDB service calls.
	//
	// The default client is created with configcache.Get.
	Client *dynamodb.Client

	// Encoder is the attributevalue.Encoder to marshal structs into DynamoDB items.
	//
	// The default Encoder is attributevalue.NewEncoder.
	Encoder *attributevalue.Encoder

	// Decoder is the attributevalue.Decoder to unmarshal attributes from DynamoDB.
	//
	// The default Decoder is attributevalue.NewDecoder.
	Decoder *attributevalue.Decoder

	// VersionUpdater is used to generate the next version value by updating the item's version value in-place.
	//
	// If VersionUpdater is given, it is always used to update the version. Otherwise, the version's Go type determines
	// how its next value is computed:
	//	- For uint and int types, the next value is version + 1.
	//	- For string and string aliases, uuid.NewString produces the next value.
	//	- For any other types, VersionUpdater must be explicitly provided.
	//
	// The function is given the exact same item passed into the operations that support optimistic locking.
	VersionUpdater func(item any)
}
