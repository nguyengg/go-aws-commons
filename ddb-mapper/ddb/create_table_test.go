package ddb_test

import (
	"testing"

	"github.com/nguyengg/go-aws-commons/ddb-mapper"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/ddb"
	. "github.com/nguyengg/go-aws-commons/ddb-mapper/internal/ddb-local-test"
	"github.com/stretchr/testify/require"
)

func TestCreateTable(t *testing.T) {
	type Item struct {
		ID string `dynamodbav:"id,hashkey" tableName:"Items"`
	}

	client := Setup(t)
	ddb.DefaultClientProvider = ddb.StaticClientProvider{Client: client}

	require.NoError(t, ddb.CreateTable(t.Context(), Item{}))
}

func TestMapper_CreateTable(t *testing.T) {
	client := Setup(t)

	type Item struct {
		ID string `dynamodbav:"id,hashkey" tableName:"Items"`
	}

	m, err := mapper.New[Item](func(m *mapper.Mapper[Item]) {
		m.Client = client
	})
	require.NoError(t, err)
	require.NoError(t, m.CreateTable(t.Context()))
}
