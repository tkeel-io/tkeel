package service

import (
	"context"

	pb "github.com/tkeel-io/tkeel/api/entry/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type EntryService struct {
	pb.UnimplementedEntryServer
}

func NewEntryService() *EntryService {
	return &EntryService{}
}

func (s *EntryService) GetEntries(ctx context.Context, req *emptypb.Empty) (*pb.GetEntriesResponse, error) {
	return &pb.GetEntriesResponse{
		Entries: []*pb.EntryObject{
			{
				Id:    "aaa",
				Name:  "aaa manager",
				Path:  "/users",
				Entry: "https://tkeel-console-plugin-users.pek3b.qingstor.com/index.html",
				Menu:  []string{"aaa"},
			},
			{
				Id:    "bbb",
				Name:  "bbb manager",
				Path:  "/plugins",
				Entry: "https://tkeel-console-plugin-plugins.pek3b.qingstor.com/index.html",
				Menu:  []string{"bbb"},
			},
		},
	}, nil
}
