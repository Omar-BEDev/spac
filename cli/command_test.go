/*
Copyright 2026 Omar-BEDev

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cli

import (
	"testing"
)

func TestIsAvailableCommand(t *testing.T) {
	tests := []struct {
		name     string
		excepted bool
	}{
		{
			name:     "post",
			excepted: true,
		},
		{
			name:     "get",
			excepted: true,
		},
		{
			name:     "put",
			excepted: true,
		},
		{
			name:     "delete",
			excepted: true,
		},
		{
			name:     "new Req",
			excepted: true,
		},
		{
			name:     "",
			excepted: false,
		},
		{
			name:     "POST",
			excepted: false,
		},
		{
			name:     " post",
			excepted: false,
		},
		{
			name:     "post ",
			excepted: false,
		},
		{
			name:     "new req",
			excepted: false,
		},
		{
			name:     "new Req ",
			excepted: false,
		},
		{
			name:     "gert",
			excepted: false,
		},
		{
			name:     "api",
			excepted: false,
		},
		{
			name:     "method",
			excepted: false,
		},
		{
			name:     "cmd",
			excepted: false,
		},
	}
	for _, tt := range tests {
		result := IsAvailableCommand(tt.name)
		if result != tt.excepted {
			t.Errorf("IsAvailableCommand(%q) = %v ; excepted (%v)", tt.name, result, tt.excepted)
		}
	}
}
