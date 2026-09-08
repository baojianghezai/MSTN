package upload

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

var allowedUploadExtensions = map[string]struct{}{
	".jpg": {}, ".jpeg": {}, ".png": {}, ".gif": {}, ".webp": {}, ".bmp": {}, ".ico": {}, ".avif": {},
	".mp3": {}, ".wav": {}, ".ogg": {}, ".m4a": {}, ".flac": {}, ".aac": {},
	".mp4": {}, ".webm": {}, ".mov": {}, ".avi": {}, ".mkv": {},
	".txt": {}, ".md": {}, ".csv": {}, ".json": {}, ".log": {}, ".pdf": {},
	".doc": {}, ".docx": {}, ".xls": {}, ".xlsx": {}, ".ppt": {}, ".pptx": {},
	".zip": {}, ".rar": {}, ".7z": {}, ".tar": {}, ".gz": {}, ".tgz": {}, ".bin": {},
}

var inlineUploadExtensions = map[string]struct{}{
	".jpg": {}, ".jpeg": {}, ".png": {}, ".gif": {}, ".webp": {}, ".bmp": {}, ".ico": {}, ".avif": {},
	".mp3": {}, ".wav": {}, ".ogg": {}, ".m4a": {}, ".flac": {}, ".aac": {},
	".mp4": {}, ".webm": {}, ".mov": {}, ".avi": {}, ".mkv": {},
}

func validatedExtension(filename string) (string, error) {
	if filename == "" || strings.TrimSpace(filename) != filename || strings.ContainsAny(filename, `/\`) {
		return "", errors.New("file extension is not allowed")
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if _, ok := allowedUploadExtensions[ext]; !ok {
		return "", errors.New("file extension is not allowed")
	}
	return ext, nil
}

// ValidateFileExtension rejects active content and unknown upload types.
func ValidateFileExtension(filename string) error {
	_, err := validatedExtension(filename)
	return err
}

// CanServeInline limits inline responses to raster images and audio/video files.
func CanServeInline(filename string) bool {
	ext, err := validatedExtension(filename)
	if err != nil {
		return false
	}
	_, ok := inlineUploadExtensions[ext]
	return ok
}

// ValidateRecruitmentFile restricts the public recruitment upload API to the
// formats used by logos, banners, resumes, and text attachments. It validates
// both the filename and the detected file signature before storage.
func ValidateRecruitmentFile(file *multipart.FileHeader, maxSize int64) error {
	if file == nil || file.Size <= 0 {
		return errors.New("upload file is empty")
	}
	if maxSize > 0 && file.Size > maxSize {
		return fmt.Errorf("upload file exceeds the %d byte limit", maxSize)
	}
	if err := ValidateFileExtension(file.Filename); err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isRecruitmentExtension(ext) {
		return errors.New("file type is not allowed for recruitment uploads")
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	head := make([]byte, 512)
	n, err := io.ReadFull(src, head)
	if err != nil && err != io.ErrUnexpectedEOF {
		return err
	}
	contentType := http.DetectContentType(head[:n])
	if !matchesRecruitmentContent(ext, contentType) {
		return errors.New("file content does not match its extension")
	}
	return nil
}

func isRecruitmentExtension(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf", ".txt", ".md", ".csv", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx":
		return true
	default:
		return false
	}
}

func matchesRecruitmentContent(ext, contentType string) bool {
	switch ext {
	case ".jpg", ".jpeg":
		return contentType == "image/jpeg"
	case ".png":
		return contentType == "image/png"
	case ".gif":
		return contentType == "image/gif"
	case ".webp":
		return contentType == "image/webp"
	case ".pdf":
		return contentType == "application/pdf"
	case ".doc":
		return contentType == "application/msword" || contentType == "application/octet-stream"
	case ".docx", ".xlsx", ".pptx":
		return contentType == "application/zip" || contentType == "application/octet-stream"
	case ".xls", ".ppt":
		return contentType == "application/vnd.ms-excel" || contentType == "application/vnd.ms-powerpoint" || contentType == "application/octet-stream"
	case ".txt", ".md", ".csv":
		return strings.HasPrefix(contentType, "text/") || contentType == "application/octet-stream"
	default:
		return false
	}
}
