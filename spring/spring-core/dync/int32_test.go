/*
 * Copyright 2012-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dync_test

import (
	"testing"

	"github.com/go-spring/spring-base/assert"
	"github.com/go-spring/spring-base/json"
	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/dync"
)

func TestInt32(t *testing.T) {

	var u dync.Int32
	assert.Equal(t, u.Value(), int32(0))

	param := conf.BindParam{
		Key:  "int",
		Path: "int32",
		Tag: conf.ParsedTag{
			Key: "int",
		},
	}

	p := conf.Map(nil)
	err := u.Validate(p, param)
	assert.Error(t, err, "bind int32 error; .* resolve property \"int\" error; property \"int\" not exist")
	err = u.Refresh(p, param)
	assert.Error(t, err, "bind int32 error; .* resolve property \"int\" error; property \"int\" not exist")

	_ = p.Set("int", int32(3))
	err = u.Validate(p, param)
	assert.Nil(t, err)
	assert.Equal(t, u.Value(), int32(0))

	param.Validate = "$>5"
	err = u.Validate(p, param)
	assert.Error(t, err, "validate failed on \"\\$\\>5\" for value 3")
	err = u.Refresh(p, param)
	assert.Error(t, err, "validate failed on \"\\$\\>5\" for value 3")

	param.Validate = ""
	err = u.Refresh(p, param)
	assert.Equal(t, u.Value(), int32(3))

	b, err := json.Marshal(&u)
	assert.Nil(t, err)
	assert.Equal(t, string(b), "3")
}
