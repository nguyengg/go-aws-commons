package keys

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// FromParameterStore creates a new [Provider] backed by AWS Systems Manager Parameter Store.
func FromParameterStore(client GetParameterAPIClient, name string, optFns ...func(opts *ParameterStoreOptions)) Provider {
	opts := &ParameterStoreOptions{}
	for _, fn := range optFns {
		fn(opts)
	}

	if opts.ValueDecoder == nil {
		opts.ValueDecoder = decodeString
	}

	return &parameterStoreKeyProvider{
		client:         client,
		name:           name,
		withDecryption: opts.WithDecryption,
		label:          opts.Label,
		valueDecoder:   opts.ValueDecoder,
	}
}

// GetParameterAPIClient abstracts the AWS Systems Manager API [ssm.Client.GetParameter].
type GetParameterAPIClient interface {
	GetParameter(context.Context, *ssm.GetParameterInput, ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

// ParameterStoreOptions customises NewParameterStoreEndec.
type ParameterStoreOptions struct {
	// WithDecryption overrides [ssm.GetParameterInput.WithDecryption].
	WithDecryption *bool

	// Label is added as ":label" suffix to [ssm.GetParameterInput.Name].
	//
	// If version is available, the label will not be used since Parameter Store (unlike Secrets Manager) only allows
	// specifying one or the other.
	Label *string

	// ValueDecoder can be used to control how the parameter value is decoded into a key.
	//
	// If not given, the default function will try these two in order:
	//  1. base64.StdEncoding.DecodeString
	//  2. hex.DecodeString
	ValueDecoder func(string) ([]byte, error)
}

type parameterStoreKeyProvider struct {
	client         GetParameterAPIClient
	name           string
	withDecryption *bool
	label          *string
	valueDecoder   func(string) ([]byte, error)
}

func (p parameterStoreKeyProvider) Provide(ctx context.Context, version *string) ([]byte, *string, error) {
	var name string
	if version != nil {
		name = fmt.Sprintf("%s:%s", p.name, *version)
	} else if p.label != nil {
		name = fmt.Sprintf("%s:%s", p.name, *p.label)
	} else {
		name = p.name
	}

	getParameterOutput, err := p.client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           &name,
		WithDecryption: p.withDecryption,
	})
	if err != nil {
		return nil, nil, err
	}

	param := getParameterOutput.Parameter
	key, err := p.valueDecoder(aws.ToString(param.Value))
	return key, aws.String(strconv.FormatInt(param.Version, 10)), nil
}
