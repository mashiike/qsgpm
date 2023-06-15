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
 * implemented the ListGroupMemberships one by referring to the ListAnalyses paginator.
 * This is a temporary solution.
 */

// ListGroupMembershipsAPIClient is a client that implements the ListGroupMemberships operation.
type ListGroupMembershipsAPIClient interface {
	ListGroupMemberships(context.Context, *quicksight.ListGroupMembershipsInput, ...func(*quicksight.Options)) (*quicksight.ListGroupMembershipsOutput, error)
}

// ListGroupMembershipsPaginatorOptions is the paginator options for ListGroupMemberships
type ListGroupMembershipsPaginatorOptions struct {
	// The maximum number of results to return.
	MaxResults *int32

	// Set to true if pagination should stop if the service returns a pagination token
	// that matches the most recent token provided to the service.
	StopOnDuplicateToken bool
}

// ListGroupMembershipsPaginator is a paginator for ListGroupMemberships
type ListGroupMembershipsPaginator struct {
	options   ListGroupMembershipsPaginatorOptions
	client    ListGroupMembershipsAPIClient
	params    *quicksight.ListGroupMembershipsInput
	nextToken *string
	firstPage bool
}

// NewListGroupMembershipsPaginator returns a new ListGroupMembershipsPaginator
func NewListGroupMembershipsPaginator(client ListGroupMembershipsAPIClient, params *quicksight.ListGroupMembershipsInput, optFns ...func(*ListGroupMembershipsPaginatorOptions)) *ListGroupMembershipsPaginator {
	if params == nil {
		params = &quicksight.ListGroupMembershipsInput{}
	}

	options := ListGroupMembershipsPaginatorOptions{}
	options.MaxResults = params.MaxResults

	for _, fn := range optFns {
		fn(&options)
	}

	return &ListGroupMembershipsPaginator{
		options:   options,
		client:    client,
		params:    params,
		firstPage: true,
		nextToken: params.NextToken,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *ListGroupMembershipsPaginator) HasMorePages() bool {
	return p.firstPage || (p.nextToken != nil && len(*p.nextToken) != 0)
}

// NextPage retrieves the next ListGroupMemberships page.
func (p *ListGroupMembershipsPaginator) NextPage(ctx context.Context, optFns ...func(*quicksight.Options)) (*quicksight.ListGroupMembershipsOutput, error) {
	if !p.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	params := *p.params
	params.NextToken = p.nextToken

	params.MaxResults = p.options.MaxResults

	result, err := p.client.ListGroupMemberships(ctx, &params, optFns...)
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
