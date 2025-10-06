package repository

import (
    `quickflow/shared/models`
)

type FileBuilder struct {
    file *models.File
}

func NewFileBuilder() *FileBuilder {
    return &FileBuilder{file: &models.File{}}
}

func (b *FileBuilder) WithName(name string) *FileBuilder {
    b.file.Name = name
    return b
}

func (b *FileBuilder) WithURL(url string) *FileBuilder {
    b.file.URL = url
    return b
}

func (b *FileBuilder) WithSize(size int64) *FileBuilder {
    b.file.Size = size
    return b
}

func (b *FileBuilder) WithExt(ext string) *FileBuilder {
    b.file.Ext = ext
    return b
}

func (b *FileBuilder) WithDisplayType(displayType models.DisplayType) *FileBuilder {
    b.file.DisplayType = displayType
    return b
}

func (b *FileBuilder) Build() *models.File {
    return b.file
}

func MotherFile1() *models.File {
    return NewFileBuilder().WithName("file1").WithURL("http://example.com/file1").Build()
}

func MotherFile2() *models.File {
    return NewFileBuilder().WithName("file2").WithURL("http://example.com/file2").Build()
}
