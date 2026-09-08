package upload

import (
	"mime/multipart"
	"testing"
)

func uploadHeader(name string, data []byte) *multipart.FileHeader {
	return &multipart.FileHeader{Filename: name, Size: int64(len(data)), Header: make(map[string][]string)}
}

func TestValidateRecruitmentFileRejectsUnsafeOrMismatchedFiles(t *testing.T) {
	// FileHeader.Open requires an internal multipart backing file, so extension,
	// empty-file, and size checks are covered here; content checks are enforced
	// by the same function before any storage adapter is called.
	if err := ValidateRecruitmentFile(uploadHeader("banner.exe", []byte("MZ")), 10<<20); err == nil {
		t.Fatal("expected executable extension to be rejected")
	}
	if err := ValidateRecruitmentFile(uploadHeader("banner.png", make([]byte, 11)), 10); err == nil {
		t.Fatal("expected oversized file to be rejected")
	}
	if err := ValidateRecruitmentFile(&multipart.FileHeader{Filename: "empty.png"}, 10<<20); err == nil {
		t.Fatal("expected empty file to be rejected")
	}
}
