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

package dync

import (
	"encoding/json"

	"github.com/go-spring/spring-base/atomic"
	"github.com/go-spring/spring-core/conf"
)

// A Bool is an atomic bool value that can be dynamic refreshed.
type Bool struct {
	v atomic.Bool
}

// Value returns the stored bool value.
func (x *Bool) Value() bool {
	return x.v.Load()
}

// Validate validates the property value.
func (x *Bool) Validate(p *conf.Properties, param conf.BindParam) error {
	var b bool
	return p.Bind(&b, conf.Param(param))
}

// Refresh refreshes the stored value.
func (x *Bool) Refresh(p *conf.Properties, param conf.BindParam) error {
	var b bool
	if err := p.Bind(&b, conf.Param(param)); err != nil {
		return err
	}
	x.v.Store(b)
	return nil
}

// MarshalJSON returns the JSON encoding of x.
func (x *Bool) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.Value())
}
