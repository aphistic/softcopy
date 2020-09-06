package fs

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	"github.com/google/uuid"
	"google.golang.org/grpc"

	"github.com/aphistic/softcopy/internal/pkg/logging"
	scproto "github.com/aphistic/softcopy/pkg/proto"
)

var reFullFilenameParts = regexp.MustCompile(`(\d+-\d+-\d+)-(.+)`)

func getFullFilename(docDate time.Time, filename string) string {
	return fmt.Sprintf(
		"%s-%s",
		docDate.Format("2006-01-02"),
		filename,
	)
}

func splitFullFilename(fullName string) (time.Time, string, error) {
	rawParts := reFullFilenameParts.FindAllStringSubmatch(fullName, -1)
	if len(rawParts) != 1 || len(rawParts[0]) != 3 {
		return time.Time{}, "", fmt.Errorf("invalid filename format")
	}

	parts := rawParts[0]

	date, err := time.Parse("2006-01-02", parts[1])
	if err != nil {
		return time.Time{}, "", err
	}

	return date, parts[2], nil
}

type FileSystemOption func(*FileSystem)

func WithLogger(logger logging.Logger) FileSystemOption {
	return func(fs *FileSystem) {
		fs.logger = logger
	}
}

type FileSystem struct {
	logger logging.Logger
	client scproto.SoftcopyClient

	inodeLock sync.RWMutex
	inodeToID map[uint64]uuid.UUID
	idToInode map[uuid.UUID]uint64
	nextInode uint64
}

func NewFileSystem(host string, port int, opts ...FileSystemOption) (*FileSystem, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := scproto.NewSoftcopyClient(conn)

	fs := &FileSystem{
		logger: logging.NewNilLogger(),
		client: client,

		inodeToID: map[uint64]uuid.UUID{},
		idToInode: map[uuid.UUID]uint64{},
	}

	for _, opt := range opts {
		opt(fs)
	}

	// fuse.Debug = func(msg interface{}) {
	// 	fs.logger.Debug("bazil debug: %v", msg)
	// }

	return fs, nil
}

func (f *FileSystem) Root() (fusefs.Node, error) {
	return newFSRootDir(f), nil
}

func (f *FileSystem) Statfs(
	ctx context.Context,
	req *fuse.StatfsRequest,
	res *fuse.StatfsResponse,
) error {
	res.Namelen = 255
	return nil
}

func (f *FileSystem) inodeForID(id uuid.UUID) uint64 {
	f.inodeLock.RLock()
	inode, ok := f.idToInode[id]
	f.inodeLock.RUnlock()
	if ok {
		return inode
	}

	f.inodeLock.Lock()
	inode, ok = f.idToInode[id]
	if ok {
		f.inodeLock.Unlock()
		return inode
	}

	inode = atomic.AddUint64(&f.nextInode, 1)
	f.inodeToID[inode] = id
	f.idToInode[id] = inode

	f.inodeLock.Unlock()

	return inode
}
