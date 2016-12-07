// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package syslog

import (
	"fmt"
	"io"
	"log/syslog"
	"os"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
)

//---------------------------------------------------------------------

// WriterI is an interface for writing a Message to some sort of output.
type WriterI interface {
	Write(*Message) error
}

//---------------------------------------------------------------------

// Writer implements the WriterI interface, writing to a generic "io.Writer" target
type Writer struct {
	Writer io.Writer
}

// Write writes the message to the io.Writer supplied.
func (w *Writer) Write(mssg *Message) error {
	if w == nil || w.Writer == nil {
		return fmt.Errorf("writer not set not set")
	}

	s := mssg.String()
	_, err := io.WriteString(w.Writer, s)
	return err
}

//---------------------------------------------------------------------

// FileWriter implements the WriterI, writing to a given file
type FileWriter struct {
	FileName string
	file     *os.File
}

// Write writes the message to the supplied file.
func (w *FileWriter) Write(mssg *Message) error {
	var err error

	if w == nil || w.FileName == "" {
		return fmt.Errorf("writer not set not set")
	}

	if w.file == nil {
		w.file, err = os.OpenFile(w.FileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
		if err != nil {
			return err
		}
	}

	s := mssg.String()
	s += "\n"

	_, err = io.WriteString(w.file, s)
	return err
}

// Close closes the file. The creator of the FileWriter must call this.
func (w *FileWriter) Close() error {
	return w.file.Close()
}

//---------------------------------------------------------------------

//ElasticWriter implements the WriterI, writing to elasticsearch
type ElasticWriter struct {
	Esi elasticsearch.IIndex
	typ string
	id  string
}

//Write writes the message to the elasticsearch index, type, id
func (w *ElasticWriter) Write(mssg *Message) error {
	var err error

	if w == nil || w.Esi == nil || w.typ == "" {
		return fmt.Errorf("writer not set not set")
	}

	_, err = w.Esi.PostData(w.typ, w.id, mssg)
	return err
}

//SetType sets the type to write to
func (w *ElasticWriter) SetType(typ string) error {
	if w == nil {
		return fmt.Errorf("writer not set not set")
	}
	w.typ = typ
	return nil
}

//SetID sets the id to write to
func (w *ElasticWriter) SetID(id string) error {
	if w == nil {
		return fmt.Errorf("writer not set not set")
	}
	w.id = id
	return nil
}

//---------------------------------------------------------------------

//SyslogWriter implements the WriterI, writing to syslog
type SyslogWriter struct {
	Syslog *syslog.Writer
}

//Write writes the message to syslog
func (w *SyslogWriter) Write(mssg *Message) error {
	var err error

	if w == nil || w.Syslog == nil {
		return fmt.Errorf("writer not set not set")
	}

	_, err = w.Syslog.Write([]byte(mssg.String()))
	return err
}
