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

package core

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/go-spring/spring-core/log"
	"github.com/go-spring/spring-core/util"
)

type beanAssembly interface {

	// Matches 成功返回 true，失败返回 false
	Matches(cond Condition) bool

	// BindStructField 对结构体的字段进行属性绑定
	BindStructField(v reflect.Value, str string, opt BindOption) error

	// WireStructField 对结构体的字段进行绑定
	WireStructField(v reflect.Value, tag string, parent reflect.Value, field string)
}

type Arg interface{}

type ArgList struct {
	fnType       reflect.Type
	withReceiver bool
	args         []Arg
}

func NewArgList(fnType reflect.Type, withReceiver bool, args []Arg) *ArgList {
	return &ArgList{fnType: fnType, withReceiver: withReceiver, args: args}
}

// Get 获取函数参数的绑定值，fileLine 是函数所在文件及其行号，日志使用
func (argList *ArgList) Get(assembly beanAssembly, fileLine string) []reflect.Value {

	fnType := argList.fnType
	numIn := fnType.NumIn()

	// 第一个参数是接收者
	if argList.withReceiver {
		numIn -= 1
	}

	variadic := fnType.IsVariadic()
	result := make([]reflect.Value, 0)

	for i, arg := range argList.args {
		var it reflect.Type

		if variadic && i >= numIn-1 {
			if argList.withReceiver {
				it = fnType.In(numIn)
			} else {
				it = fnType.In(numIn - 1)
			}
		} else {
			if argList.withReceiver {
				it = fnType.In(i + 1)
			} else {
				it = fnType.In(i)
			}
		}

		if variadic && i >= numIn-1 { // 可变参数
			ev := argList.getArgValue(it.Elem(), arg, assembly, fileLine)
			if ev.IsValid() {
				result = append(result, ev)
			}
		} else {
			iv := argList.getArgValue(it, arg, assembly, fileLine)
			result = append(result, iv)
		}
	}

	return result
}

// getArgValue 获取绑定参数值
func (argList *ArgList) getArgValue(t reflect.Type, arg Arg, assembly beanAssembly, fileLine string) reflect.Value {

	description := fmt.Sprintf("arg:\"%v\" %s", arg, fileLine)
	defer log.Tracef("get value success %s", description)
	log.Tracef("get value %s", description)

	switch tArg := arg.(type) {
	case string:
		{
			v := reflect.New(t).Elem()
			if IsValueType(v.Kind()) { // 值类型，采用属性绑定语法
				if tArg == "" {
					tArg = "${}"
				}
				err := assembly.BindStructField(v, tArg, BindOption{})
				util.Panic(err).When(err != nil)
			} else { // 引用类型，采用对象注入语法
				assembly.WireStructField(v, tArg, reflect.Value{}, "")
			}
			return v
		}
	case *OptionArg:
		{
			return tArg.call(assembly)
		}
	}
	panic("getArgValue:::")
}

//type ValueArg struct {
//}

// OptionArg Option 函数的绑定参数
type OptionArg struct {
	cond Condition // 判断条件

	fn      interface{}
	argList *ArgList

	file string // 注册点所在文件
	line int    // 注册点所在行数
}

// 判断是否是合法的 Option 函数，只能有一个返回值
func validOptionFunc(fnType reflect.Type) bool {
	return fnType.Kind() == reflect.Func && fnType.NumOut() == 1
}

// NewOptionArg OptionArg 的构造函数，args 是 Option 函数的一般参数绑定
func NewOptionArg(fn interface{}, strArgs ...string) *OptionArg {

	var (
		file string
		line int
	)

	// 获取注册点信息
	for i := 1; i < 10; i++ {
		_, file0, line0, _ := runtime.Caller(i)

		// 排除 spring-core 包下面所有的非 test 文件
		if strings.Contains(file0, "/spring-core/") {
			if !strings.HasSuffix(file0, "_test.go") {
				continue
			}
		}

		file = file0
		line = line0
		break
	}

	fnType := reflect.TypeOf(fn)
	if ok := validOptionFunc(fnType); !ok {
		panic(errors.New("option func must be func(...)option"))
	}

	args := make([]Arg, len(strArgs))
	for i, arg := range strArgs {
		args[i] = arg
	}

	return &OptionArg{
		fn:      fn,
		argList: NewArgList(fnType, false, args),
		file:    file,
		line:    line,
	}
}

func (arg *OptionArg) FileLine() string {
	return fmt.Sprintf("%s:%d", arg.file, arg.line)
}

// WithCondition 为 OptionArg 设置一个 Condition
func (arg *OptionArg) WithCondition(cond Condition) *OptionArg {
	arg.cond = cond
	return arg
}

// call 获取 OptionArg 的运算值
func (arg *OptionArg) call(assembly beanAssembly) reflect.Value {

	defer log.Tracef("call option func success %s", arg.FileLine())
	log.Tracef("call option func %s", arg.FileLine())

	if arg.cond == nil || assembly.Matches(arg.cond) {
		fnValue := reflect.ValueOf(arg.fn)
		in := arg.argList.Get(assembly, arg.FileLine())
		out := fnValue.Call(in)
		return out[0]
	}

	return reflect.Value{}
}
