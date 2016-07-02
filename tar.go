package main

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Mode constants from the tar spec.
const (
	tar_ISDIR = 040000  // Directory
	tar_ISREG = 0100000 // Regular file
	tar_ISLNK = 0120000 // Symbolic link
)

func PackTarWalkFunc(w *tar.Writer, rel string) filepath.WalkFunc {
	return func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		hdr := tar.Header{
			Name:  path,
			Mode:  0644,
			Uname: "root",
			Gname: "root",
		}

		if rel != "" {
			if hdr.Name, err = filepath.Rel(rel, hdr.Name); err != nil {
				return err
			}
		}

		fm := fi.Mode()
		switch {
		default:
			return &os.PathError{
				Op:   "pack",
				Path: path,
				Err:  fmt.Errorf("impossible file mode: %s", fm),
			}
		case fm.IsDir():
			hdr.Typeflag = tar.TypeDir
			hdr.Mode |= tar_ISDIR | 0111
		case fm&os.ModeSymlink != 0:
			hdr.Typeflag = tar.TypeSymlink
			hdr.Mode |= tar_ISLNK | 0777
			if hdr.Linkname, err = os.Readlink(path); err != nil {
				return err
			}
		case fm.IsRegular():
			hdr.Typeflag = tar.TypeReg
			hdr.Mode |= tar_ISREG
			if fm&0111 != 0 {
				hdr.Mode |= 0111
			}
			hdr.Size = fi.Size()
		}

		if err := w.WriteHeader(&hdr); err != nil {
			return err
		}

		if fm.IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(w, f); err != nil {
				return err
			}
		}

		return nil
	}
}

func PackTarForestWalkFunc(w *tar.Writer, rel string) filepath.WalkFunc {
	return func(path string, fi os.FileInfo, err error) error {
		hdr := tar.Header{
			Name:  filepath.ToSlash(path),
			Mode:  0644,
			Uname: "root",
			Gname: "root",
		}
		if rel != "" {
			if hdr.Name, err = filepath.Rel(rel, hdr.Name); err != nil {
				return err
			}
		}

		fm := fi.Mode()
		switch {
		case fm.IsDir():
			hdr.Typeflag = tar.TypeDir
			hdr.Mode |= tar_ISDIR | 0111
		default:
			hdr.Typeflag = tar.TypeSymlink
			hdr.Mode |= tar_ISLNK | 0777
			hdr.Linkname = path
		}

		return w.WriteHeader(&hdr)
	}
}
