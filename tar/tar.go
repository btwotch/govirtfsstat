package tar

import (
	"archive/tar"
	"errors"
	"fmt"
	"govirtfsstat/stat"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

func Untar(tarReader io.Reader, destDir string) error {
	tr := tar.NewReader(tarReader)

	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return fmt.Errorf("extracting tar header failed: %w", err)
		}

		err = UntarEntry(destDir, hdr, tr)
		if err != nil {
			return fmt.Errorf("untaring failed: %w", err)
		}
	}

	return nil
}

func UntarEntry(destDir string, hdr *tar.Header, tr *tar.Reader) error {
	path := filepath.Join(destDir, hdr.Name)
	if !strings.HasPrefix(path, destDir) {
		path = filepath.Join(destDir, filepath.Base(hdr.Name))
	}

	return UntarEntryPath(path, hdr, tr)
}

func UntarEntryPath(path string, hdr *tar.Header, tr *tar.Reader) error {
	mode := hdr.FileInfo().Mode()

	if hdr.Typeflag == tar.TypeDir || hdr.FileInfo().IsDir() {
		err := os.MkdirAll(path, 0700)
		if err != nil {
			return fmt.Errorf("mkdir '%s' failed: %w", path, err)
		}
		mode = mode | syscall.S_IFDIR
	} else {
		err := os.MkdirAll(filepath.Dir(path), 0700)
		if err != nil {
			return fmt.Errorf("mkdir '%s' for '%s' failed: %w", filepath.Dir(path), path, err)
		}
		fp, err := os.OpenFile(path, syscall.O_CREAT|syscall.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("open '%s' failed: %w", path, err)
		}
		_, err = io.Copy(fp, tr)
		if err != nil {
			return fmt.Errorf("copy failed: %w", err)
		}
		fp.Close()
	}

	if hdr.Typeflag == tar.TypeBlock || hdr.Typeflag == tar.TypeChar {
		dev := unix.Mkdev(uint32(hdr.Devmajor), uint32(hdr.Devminor))
		err := stat.SetRDev(path, dev)
		if err != nil {
			return fmt.Errorf("set rdev failed: %w", err)
		}
	}

	if hdr.Typeflag == tar.TypeChar {
		mode = mode | syscall.S_IFCHR
	}
	if hdr.Typeflag == tar.TypeBlock {
		mode = mode | syscall.S_IFBLK
	}
	if hdr.Typeflag == tar.TypeLink {
		mode = mode | syscall.S_IFREG
	}
	if hdr.Typeflag == tar.TypeFifo {
		mode = mode | syscall.S_IFIFO
	}
	if hdr.Typeflag == tar.TypeSymlink {
		mode = mode | syscall.S_IFLNK
		err := os.WriteFile(path, []byte(hdr.Linkname), 0600)
		if err != nil {
			panic(err)
		}
	}

	err := stat.SetMode(path, uint32(mode))
	if err != nil {
		return fmt.Errorf("set mode failed: %w", err)
	}
	err = stat.SetUid(path, uint32(hdr.Uid))
	if err != nil {
		return fmt.Errorf("set uid failed: %w", err)
	}
	err = stat.SetGid(path, uint32(hdr.Gid))
	if err != nil {
		return fmt.Errorf("set gid failed: %w", err)
	}

	return nil
}
