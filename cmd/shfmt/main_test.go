// Copyright (c) 2016, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

var walkTests = []struct {
	run        bool
	name, body string
}{
	{true, "shebang-1", "#!/bin/sh\n foo"},
	{true, "shebang-2", "#!/bin/bash\n foo"},
	{true, "shebang-3", "#!/usr/bin/sh\n foo"},
	{true, "shebang-4", "#!/usr/bin/env bash\n foo"},
	{true, "shebang-5", "#!/bin/env sh\n foo"},
	{true, "ext.sh", " foo"},
	{true, "ext.bash", " foo"},
	{true, "ext-shebang.sh", "#!/bin/sh\n foo"},
	{false, ".hidden", " foo long enough"},
	{false, ".hidden-shebang", "#!/bin/sh\n foo"},
	{false, "noext-empty", " foo"},
	{false, "noext-noshebang", " foo long enough"},
	{false, "ext.other", " foo"},
	{false, "ext-shebang.other", "#!/bin/sh\n foo"},
}

func TestWalk(t *testing.T) {
	dir, err := ioutil.TempDir("", "shfmt-walk")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	for _, wt := range walkTests {
		if err := ioutil.WriteFile(wt.name, []byte(wt.body), 0666); err != nil {
			t.Fatal(err)
		}
	}
	var buf bytes.Buffer
	out = &buf
	*list, *write = true, true
	onError := func(err error) {
	}
	if err := walk(".", onError); err != nil {
		t.Fatal(err)
	}
	modified := make(map[string]bool, 0)
	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		name := scanner.Text()
		modified[name] = true
	}
	for _, wt := range walkTests {
		t.Run(wt.name, func(t *testing.T) {
			if modified[wt.name] == wt.run {
				return
			}
			if wt.run {
				t.Fatalf("walk had to run on %s but didn't", wt.name)
			} else {
				t.Fatalf("walk had to not run on %s but did", wt.name)
			}
		})
	}
	if err := walk(".", onError); err != nil {
		t.Fatal(err)
	}
	if buf.Len() > 0 {
		t.Fatal("shfmt -l -w printed filenames on a duplicate run")
	}
	*list, *write = false, false
	if err := walk(".", onError); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("shfmt without -l nor -w did not print anything")
	}
}