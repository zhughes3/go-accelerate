package server

import (
	"context"
	"github.com/zhughes3/go-accelerate/internal/pkg/endpoint"
	achttp "github.com/zhughes3/go-accelerate/internal/pkg/http"
	"github.com/zhughes3/go-accelerate/internal/pkg/v1/api"
	"github.com/zhughes3/go-accelerate/internal/pkg/v1/user"
	"github.com/zhughes3/go-accelerate/pkg/timelines"
	"github.com/zhughes3/go-accelerate/pkg/url"
	"net/http"
)

const pathV1 = "/v1/api"

const (
	paramTimelineID  = "timeline_id"
	pathTimelineList = "timelines"
)

func NewHandlerSpecs(userID user.IDResolver, timelinesService timelines.Service) []achttp.RequestHandlerSpec {
	return NewTimelineHandlerSpecs(pathV1, timelinesService, userID)
}

func NewTimelineHandlerSpecs(pathPrefix string, s timelines.Service, userID user.IDResolver) []achttp.RequestHandlerSpec {
	return []achttp.RequestHandlerSpec{
		newCreateTimelineHandlerSpec(pathPrefix, s, userID),
		newListTimelinesHandlerSpec(pathPrefix, s, userID),
	}
}

func newListTimelinesHandlerSpec(pathPrefix string, s timelines.Service, userID user.IDResolver) achttp.RequestHandlerSpec {
	return achttp.NewRequestHandlerSpec(
		"list-timelines",
		DecodeListRequest,
		NewListTimelinesEndpoint(s, userID),
		achttp.EncodeJSONResponse,
		url.CreateFullPath(pathPrefix, pathTimelineList),
		[]string{http.MethodGet},
	)
}

func newCreateTimelineHandlerSpec(pathPrefix string, s timelines.Service, userID user.IDResolver) achttp.RequestHandlerSpec {
	return achttp.NewRequestHandlerSpec(
		"create-timeline",
		DecodeCreateRequest[api.IdentifiableTimeline],
		NewCreateTimelineEndpoint(s, userID),
		achttp.EncodeJSONCreatedResponse,
		url.CreateFullPath(pathPrefix, pathTimelineList),
		[]string{http.MethodPost},
	)
}

func NewListTimelinesEndpoint(s timelines.Service, userID user.IDResolver) endpoint.Endpoint {
	return func(ctx context.Context, _ any) (any, error) {
		return s.ListTimelines(ctx, userID(ctx))
	}
}

func NewCreateTimelineEndpoint(s timelines.Service, userID user.IDResolver) endpoint.Endpoint {
	return func(ctx context.Context, r any) (any, error) {
		tr := r.(createRequest[api.Timeline])

		return s.CreateTimeline(ctx, userID(ctx), timelines.TimelineCreateReq{Name: tr.entity.Name})
	}
}

type byIDRequest struct {
	id string
}

func NewDecodeByIDRequest(paramName string) achttp.DecodeRequestFunc {
	return func(_ context.Context, r *http.Request) (any, error) {
		var (
			request byIDRequest
			err     error
		)

		if request.id, err = achttp.RequiredParamValue(r, paramName); err != nil {
			return nil, err
		}

		return request, nil
	}
}

type createRequest[T any] struct {
	entity T
}

func DecodeCreateRequest[T any](_ context.Context, r *http.Request) (any, error) {
	var request createRequest[T]
	if err := achttp.DecodeJSONRequestBody(r, &request.entity); err != nil {
		return nil, err
	}

	return request, nil
}

type listRequest struct {
	limit   int
	page    int
	orderBy string
}

func DecodeListRequest(_ context.Context, _ *http.Request) (any, error) {
	return listRequest{}, nil
}
