package quicksightx

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/quicksight"
)

/*
 * The original, original code is here; https://github.com/aws/aws-sdk-go-v2/blob/service/quicksight/v1.18.0/service/quicksight/api_op_ListAnalyses.go#L158
 * The license for the original code is here.; https://github.com/aws/aws-sdk-go-v2/blob/service/quicksight/v1.18.0/LICENSE.txt
 *
 * implemented the ListGroups one by referring to the ListAnalyses paginator.
 * This is a temporary solution.
 */

// ListGroupsAPIClient is a client that implements the ListGroups operation.
type ListGroupsAPIClient interface {
	ListGroups(context.Context, *quicksight.ListGroupsInput, ...func(*quicksight.Options)) (*quicksight.ListGroupsOutput, error)
}

// ListGroupsPaginatorOptions is the paginator options for ListGroups
type ListGroupsPaginatorOptions struct {
	// The maximum number of results to return.
	MaxResults *int32

	// Set to true if pagination should stop if the service returns a pagination token
	// that matches the most recent token provided to the service.
	StopOnDuplicateToken bool
}

// ListGroupsPaginator is a paginator for ListGroups
type ListGroupsPaginator struct {
	options   ListGroupsPaginatorOptions
	client    ListGroupsAPIClient
	params    *quicksight.ListGroupsInput
	nextToken *string
	firstPage bool
}

// NewListGroupsPaginator returns a new ListGroupsPaginator
func NewListGroupsPaginator(client ListGroupsAPIClient, params *quicksight.ListGroupsInput, optFns ...func(*ListGroupsPaginatorOptions)) *ListGroupsPaginator {
	if params == nil {
		params = &quicksight.ListGroupsInput{}
	}

	options := ListGroupsPaginatorOptions{}
	options.MaxResults = params.MaxResults

	for _, fn := range optFns {
		fn(&options)
	}

	return &ListGroupsPaginator{
		options:   options,
		client:    client,
		params:    params,
		firstPage: true,
		nextToken: params.NextToken,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *ListGroupsPaginator) HasMorePages() bool {
	return p.firstPage || (p.nextToken != nil && len(*p.nextToken) != 0)
}

// NextPage retrieves the next ListGroups page.
func (p *ListGroupsPaginator) NextPage(ctx context.Context, optFns ...func(*quicksight.Options)) (*quicksight.ListGroupsOutput, error) {
	if !p.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	params := *p.params
	params.NextToken = p.nextToken
	params.MaxResults = p.options.MaxResults

	result, err := p.client.ListGroups(ctx, &params, optFns...)
	if err != nil {
		return nil, err
	}
	p.firstPage = false

	prevToken := p.nextToken
	p.nextToken = result.NextToken

	if p.options.StopOnDuplicateToken &&
		prevToken != nil &&
		p.nextToken != nil &&
		*prevToken == *p.nextToken {
		p.nextToken = nil
	}

	return result, nil
}
