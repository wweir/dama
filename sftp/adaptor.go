package sftp

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"github.com/wweir/utils/log"
	"golang.org/x/net/webdav"
)

func (s *SSHInfo) HomePath(path string) string {
	// return path
	return filepath.Join(s.Home, path)
}

// warp for webdav
func (s *SSHInfo) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	if err := s.sftp.Mkdir(s.HomePath(name)); err != nil {
		return err
	}

	return s.sftp.Chmod(s.HomePath(name), perm)
}

type fileDir struct {
	*sftp.File
	si *SSHInfo

	dirName string
	ReadDir func(p string) ([]os.FileInfo, error)
}

func (f *fileDir) Readdir(count int) ([]os.FileInfo, error) {
	files, err := f.ReadDir(f.dirName)
	if err != nil {
		return nil, err
	}
	if len(files) != 0 {
		log.Infow("file info", "file name", files[0].Name())
	}

	// cache fileinfos
	f.si.dirCache.Store(f.dirName, files)

	if count > 0 && len(files) > count {
		return files[:count], nil
	}
	return files, nil
}

func (s *SSHInfo) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	f, err := s.sftp.OpenFile(s.HomePath(name), int(perm))
	if err != nil {
		return nil, err
	}

	return &fileDir{
		File:    f,
		si:      s,
		ReadDir: s.sftp.ReadDir,
	}, nil
}
func (s *SSHInfo) RemoveAll(ctx context.Context, name string) error {
	return s.sftp.Remove(s.HomePath(name))
}
func (s *SSHInfo) Rename(ctx context.Context, oldName, newName string) error {
	return s.sftp.Rename(s.HomePath(oldName), s.HomePath(newName))
}

// Stat with cache for speedup
func (s *SSHInfo) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	log.Infow("stat", "name", name)

	dir := filepath.Dir(s.HomePath(name))
	if name == "/" {
		dir += "/."

		if val, ok := s.dirCache.Load(dir); ok {
			return val.(os.FileInfo), nil
		}

		fi, err := s.sftp.Lstat(s.HomePath(name))
		if err != nil {
			return nil, err
		}

		s.dirCache.Store(dir, fi)
		return fi, nil
	}

	if fis, ok := s.dirCache.Load(dir); ok {
		basename := filepath.Base(name)
		for _, fi := range fis.([]os.FileInfo) {
			if basename == fi.Name() {
				return fi, nil
			}
		}
		return nil, errors.New("file does not exist: " + s.HomePath(name))
	}

	fis, err := s.sftp.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	s.dirCache.Store(dir, fis)

	return s.Stat(ctx, name)
}
