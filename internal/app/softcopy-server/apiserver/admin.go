package apiserver

import (
	"github.com/efritz/nacelle"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/aphistic/softcopy/internal/pkg/api"
	"github.com/aphistic/softcopy/internal/pkg/protoutil"
	scproto "github.com/aphistic/softcopy/pkg/proto"
)

type adminServer struct {
	logger nacelle.Logger
	api    *api.Client
}

func (as *adminServer) AllFiles(
	req *scproto.AllFileRequest,
	srv scproto.SoftcopyAdmin_AllFilesServer,
) error {
	files, err := as.api.AllFiles()
	if err != nil {
		return grpc.Errorf(codes.Internal, err.Error())
	}

	for {
		select {
		case <-srv.Context().Done():
			as.logger.Info("all files context ended")
			return nil
		case item, ok := <-files.Files():
			if !ok {
				return nil
			}

			if item.Error != nil {
				as.logger.Error("error getting file: %s", item.Error)
				continue
			}

			resFile, err := protoutil.FileToProto(item.File)
			if err != nil {
				as.logger.Error("Error getting grpc version of file: %s", err)
				continue
			}

			tags, err := as.api.GetTagsForFile(item.File.ID.String())
			if err != nil {
				as.logger.Error("Error getting tags for file: %s", err)
				continue
			}

			resTags := []*scproto.Tag{}
		tagLoop:
			for {
				select {
				case <-srv.Context().Done():
					return nil
				case tagItem, ok := <-tags.Tags():
					if !ok {
						break tagLoop
					}

					if tagItem.Error != nil {
						as.logger.Error("Error getting tag: %s", tagItem.Error)
						continue
					}

					resTag, err := tagToGrpc(tagItem.Tag)
					if err != nil {
						as.logger.Error("Error getting grpc version of tag: %s", err)
						continue
					}
					resTags = append(resTags, resTag)
				}
			}

			taggedFile := &scproto.TaggedFile{
				File: resFile,
				Tags: resTags,
			}

			err = srv.Send(taggedFile)
			if err != nil {
				as.logger.Error("Error sending tagged file: %s", err)
				continue
			}
		}
	}
}
