package utils

import (
	"archive/zip"
	"golang.org/x/exp/errors/fmt"
	"io"
	"os"
	"path/filepath"
)

func Zip(files []*os.File, dest string) error {
	d, _ := os.Create(dest)
	defer d.Close()

	w := zip.NewWriter(d)
	defer w.Close()

	for _, file := range files {
		err := doZip(file, "", w)
		if err != nil {
			return err
		}
	}
	return nil
}

func doZip(file *os.File, prefix string, zw *zip.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}

	if info.IsDir() {
		prefix = prefix + "/" + info.Name()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = doZip(f, prefix, zw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := zip.FileInfoHeader(info)
		header.Name = prefix + "/" + header.Name
		if err != nil {
			return err
		}

		header.Method = zip.Deflate

		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		if err != nil {
			return err
		}
	}
	return nil
}

func Unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open ZIP file: %v", err)
	}
	defer r.Close()

	for _, f := range r.File {
		err = os.MkdirAll(dest+"/"+filepath.Dir(f.Name), 0755) // 创建目录结构
		if err != nil {
			return fmt.Errorf("failed to create directory structure for '%s': %v", f.Name, err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to read file '%s' from ZIP archive: %v", f.Name, err)
		}
		defer rc.Close()

		outPath := filepath.Join(dest, f.Name)
		w, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("failed to create output file '%s': %v", outPath, err)
		}
		defer w.Close()

		_, err = io.Copy(w, rc)
		if err != nil {
			return fmt.Errorf("failed to copy data into output file '%s': %v", outPath, err)
		}
	}

	return nil
}
