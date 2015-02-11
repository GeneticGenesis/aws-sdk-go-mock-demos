package app

import (
	"github.com/awslabs/aws-sdk-go/service/dynamodb"
)

type DynamoDBer interface {
	Query(*dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
}

var _ DynamoDBer = (*dynamodb.DynamoDB)(nil)

// Other bits we need, but should be able to use the real ones...
// Condition            // type
// AttributeValue       // type
// ComparisonOperatorEq // const
