package sessions_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/config"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/gin-sessions"
	local "github.com/nguyengg/go-dynamodb-local"
	"github.com/stretchr/testify/require"
)

type Session struct {
	ID       string    `dynamodbav:"id,hashkey" tablename:"Sessions"`
	User     string    `dynamodbav:"user"`
	Version  int       `dynamodbav:"version,version"`
	Created  time.Time `dynamodbav:"created,createdtime"`
	Modified time.Time `dynamodbav:"modified,modifiedtime"`
}

func setup(t *testing.T, debug ...bool) (*sessions.Manager[Session], *dynamodb.Client) {
	client := local.DefaultSkippable(t)
	if len(debug) != 0 && debug[0] {
		client = dynamodb.New(client.Options(), func(options *dynamodb.Options) {
			options.ClientLogMode = aws.LogRequestWithBody | aws.LogResponseWithBody
		})
	}
	config.DefaultClientProvider = config.StaticClientProvider{Client: client}

	require.NoError(t, ddb.CreateTable(t.Context(), Session{}))

	m, err := sessions.New[Session]()
	require.NoError(t, err)

	return m, client
}
