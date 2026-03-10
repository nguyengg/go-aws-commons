package ddb_test

import (
	"testing"

	"github.com/nguyengg/go-aws-commons/ddb-mapper"
	. "github.com/nguyengg/go-aws-commons/ddb-mapper/internal/ddb-local-test"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/mapper"
	"github.com/stretchr/testify/require"
)

func TestCreateTable(t *testing.T) {
	type Item struct {
		ID string `dynamodbav:"id,hashkey" tableName:"Items"`
	}

	Setup(t)

	// if table is created using package-level CreateTable, mapper can still read it.
	require.NoError(t, ddb.CreateTable(t.Context(), Item{}))

	m, err := mapper.New[Item]()
	require.NoError(t, err)
	_, err = m.Get(t.Context(), &Item{ID: "hello"})
	require.NoError(t, err)
}

func TestMapper_CreateTable(t *testing.T) {
	Setup(t)

	type Item struct {
		ID string `dynamodbav:"id,hashkey" tableName:"Items"`
	}

	m, err := mapper.New[Item]()
	require.NoError(t, err)

	// if table is created using Mapper.CreateTable, package-level Get can still read it.
	require.NoError(t, m.CreateTable(t.Context()))

	_, err = ddb.Get(t.Context(), &Item{ID: "hello"})
	require.NoError(t, err)
}
