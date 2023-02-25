// Copyright 2020 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package shader

import (
	"fmt"
	"go/ast"
	gconstant "go/constant"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/internal/shaderir"
)

func (cs *compileState) parseType(block *block, fname string, expr ast.Expr) (shaderir.Type, bool) {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "bool":
			return shaderir.Type{Main: shaderir.Bool}, true
		case "int":
			return shaderir.Type{Main: shaderir.Int}, true
		case "float":
			return shaderir.Type{Main: shaderir.Float}, true
		case "vec2":
			return shaderir.Type{Main: shaderir.Vec2}, true
		case "vec3":
			return shaderir.Type{Main: shaderir.Vec3}, true
		case "vec4":
			return shaderir.Type{Main: shaderir.Vec4}, true
		case "ivec2":
			return shaderir.Type{Main: shaderir.IVec2}, true
		case "ivec3":
			return shaderir.Type{Main: shaderir.IVec3}, true
		case "ivec4":
			return shaderir.Type{Main: shaderir.IVec4}, true
		case "mat2":
			return shaderir.Type{Main: shaderir.Mat2}, true
		case "mat3":
			return shaderir.Type{Main: shaderir.Mat3}, true
		case "mat4":
			return shaderir.Type{Main: shaderir.Mat4}, true
		default:
			cs.addError(t.Pos(), fmt.Sprintf("unexpected type: %s", t.Name))
			return shaderir.Type{}, false
		}
	case *ast.ArrayType:
		if t.Len == nil {
			cs.addError(t.Pos(), fmt.Sprintf("array length must be specified"))
			return shaderir.Type{}, false
		}
		var length int
		if _, ok := t.Len.(*ast.Ellipsis); ok {
			length = -1 // Determine the length later.
		} else {
			exprs, _, _, ok := cs.parseExpr(block, fname, t.Len, true)
			if !ok {
				return shaderir.Type{}, false
			}
			if len(exprs) != 1 {
				cs.addError(t.Pos(), fmt.Sprintf("invalid length of array"))
				return shaderir.Type{}, false
			}
			if exprs[0].Type != shaderir.NumberExpr {
				cs.addError(t.Pos(), fmt.Sprintf("length of array must be a constant number"))
				return shaderir.Type{}, false
			}
			l, ok := gconstant.Int64Val(exprs[0].Const)
			if !ok {
				cs.addError(t.Pos(), fmt.Sprintf("length of array must be an integer"))
				return shaderir.Type{}, false
			}
			length = int(l)
		}

		elm, ok := cs.parseType(block, fname, t.Elt)
		if !ok {
			return shaderir.Type{}, false
		}
		if elm.Main == shaderir.Array {
			cs.addError(t.Pos(), fmt.Sprintf("array of array is forbidden"))
			return shaderir.Type{}, false
		}
		return shaderir.Type{
			Main:   shaderir.Array,
			Sub:    []shaderir.Type{elm},
			Length: length,
		}, true
	case *ast.StructType:
		cs.addError(t.Pos(), "struct is not implemented")
		return shaderir.Type{}, false
	default:
		cs.addError(t.Pos(), fmt.Sprintf("unepxected type: %v", t))
		return shaderir.Type{}, false
	}
}

func canBeFloatImplicitly(expr shaderir.Expr, t shaderir.Type) (shaderir.Expr, bool) {
	// TODO: For integers, should only constants be allowed?
	if t.Main == shaderir.Int {
		return expr, true
	}
	if t.Main == shaderir.Float {
		return expr, true
	}
	if expr.Const != nil {
		if expr.Const.Kind() == gconstant.Int {
			expr.Const = gconstant.ToFloat(expr.Const)
			return expr, true
		}
		if expr.Const.Kind() == gconstant.Float {
			expr.Const = gconstant.ToFloat(expr.Const)
			return expr, true
		}
	}
	return expr, false
}

func checkArgsForBoolBuiltinFunc(args []shaderir.Expr, argts []shaderir.Type) error {
	if len(args) != len(argts) {
		return fmt.Errorf("the number of arguments and types doesn't match: %d vs %d", len(args), len(argts))
	}

	if len(args) != 1 {
		return fmt.Errorf("number of bool's arguments must be 1 but %d", len(args))
	}
	if argts[0].Main == shaderir.Bool {
		return nil
	}
	if args[0].Const != nil && args[0].Const.Kind() == gconstant.Bool {
		return nil
	}
	return fmt.Errorf("invalid arguments for bool: (%s)", argts[0].String())
}

func checkArgsForIntBuiltinFunc(args []shaderir.Expr, argts []shaderir.Type) error {
	if len(args) != len(argts) {
		return fmt.Errorf("the number of arguments and types doesn't match: %d vs %d", len(args), len(argts))
	}

	if len(args) != 1 {
		return fmt.Errorf("number of int's arguments must be 1 but %d", len(args))
	}
	if argts[0].Main == shaderir.Int || argts[0].Main == shaderir.Float {
		return nil
	}
	if args[0].Const != nil && canTruncateToInteger(&args[0].Const) {
		return nil
	}
	return fmt.Errorf("invalid arguments for int: (%s)", argts[0].String())
}

func checkArgsForFloatBuiltinFunc(args []shaderir.Expr, argts []shaderir.Type) error {
	if len(args) != len(argts) {
		return fmt.Errorf("the number of arguments and types doesn't match: %d vs %d", len(args), len(argts))
	}

	if len(args) != 1 {
		return fmt.Errorf("number of float's arguments must be 1 but %d", len(args))
	}
	if arg, ok := canBeFloatImplicitly(args[0], argts[0]); ok {
		args[0] = arg
		return nil
	}
	return fmt.Errorf("invalid arguments for float: (%s)", argts[0].String())
}

func checkArgsForVec2BuiltinFunc(args []shaderir.Expr, argts []shaderir.Type) error {
	if len(args) != len(argts) {
		return fmt.Errorf("the number of arguments and types doesn't match: %d vs %d", len(args), len(argts))
	}

	switch len(args) {
	case 1:
		if arg, ok := canBeFloatImplicitly(args[0], argts[0]); ok {
			args[0] = arg
			return nil
		}
		if argts[0].IsVector() && argts[0].VectorElementCount() == 2 {
			return nil
		}
	case 2:
		arg0, ok0 := canBeFloatImplicitly(args[0], argts[0])
		arg1, ok1 := canBeFloatImplicitly(args[1], argts[1])
		if ok0 && ok1 {
			args[0], args[1] = arg0, arg1
			return nil
		}
	default:
		return fmt.Errorf("invalid number of arguments for vec2")
	}

	var str []string
	for _, t := range argts {
		str = append(str, t.String())
	}
	return fmt.Errorf("invalid arguments for vec2: (%s)", strings.Join(str, ", "))
}

func checkArgsForVec3BuiltinFunc(args []shaderir.Expr, argts []shaderir.Type) error {
	if len(args) != len(argts) {
		return fmt.Errorf("the number of arguments and types doesn't match: %d vs %d", len(args), len(argts))
	}

	switch len(args) {
	case 1:
		if arg, ok := canBeFloatImplicitly(args[0], argts[0]); ok {
			args[0] = arg
			return nil
		}
		if argts[0].IsVector() && argts[0].VectorElementCount() == 3 {
			return nil
		}
	case 2:
		if arg, ok := canBeFloatImplicitly(args[0], argts[0]); ok && argts[1].IsVector() && argts[1].VectorElementCount() == 2 {
			args[0] = arg
			return nil
		}
		if arg, ok := canBeFloatImplicitly(args[1], argts[1]); ok && argts[0].IsVector() && argts[0].VectorElementCount() == 2 {
			args[1] = arg
			return nil
		}
	case 3:
		arg0, ok0 := canBeFloatImplicitly(args[0], argts[0])
		arg1, ok1 := canBeFloatImplicitly(args[1], argts[1])
		arg2, ok2 := canBeFloatImplicitly(args[2], argts[2])
		if ok0 && ok1 && ok2 {
			args[0], args[1], args[2] = arg0, arg1, arg2
			return nil
		}
	default:
		return fmt.Errorf("invalid number of arguments for vec3")
	}

	var str []string
	for _, t := range argts {
		str = append(str, t.String())
	}
	return fmt.Errorf("invalid arguments for vec3: (%s)", strings.Join(str, ", "))
}

func checkArgsForVec4BuiltinFunc(args []shaderir.Expr, argts []shaderir.Type) error {
	if len(args) != len(argts) {
		return fmt.Errorf("the number of arguments and types doesn't match: %d vs %d", len(args), len(argts))
	}

	switch len(args) {
	case 1:
		if arg, ok := canBeFloatImplicitly(args[0], argts[0]); ok {
			args[0] = arg
			return nil
		}
		if argts[0].IsVector() && argts[0].VectorElementCount() == 4 {
			return nil
		}
	case 2:
		if arg, ok := canBeFloatImplicitly(args[0], argts[0]); ok && argts[1].IsVector() && argts[1].VectorElementCount() == 3 {
			args[0] = arg
			return nil
		}
		if argts[0].IsVector() && argts[0].VectorElementCount() == 2 && argts[1].IsVector() && argts[1].VectorElementCount() == 2 {
			return nil
		}
		if arg, ok := canBeFloatImplicitly(args[1], argts[1]); ok && argts[0].IsVector() && argts[0].VectorElementCount() == 3 {
			args[1] = arg
			return nil
		}
	case 3:
		arg0, ok0 := canBeFloatImplicitly(args[0], argts[0])
		arg1, ok1 := canBeFloatImplicitly(args[1], argts[1])
		arg2, ok2 := canBeFloatImplicitly(args[2], argts[2])
		if ok0 && ok1 && argts[2].IsVector() && argts[2].VectorElementCount() == 2 {
			args[0], args[1] = arg0, arg1
			return nil
		}
		if ok0 && argts[1].IsVector() && argts[1].VectorElementCount() == 2 && ok2 {
			args[0], args[2] = arg0, arg2
			return nil
		}
		if argts[0].IsVector() && argts[0].VectorElementCount() == 2 && ok1 && ok2 {
			args[1], args[2] = arg1, arg2
			return nil
		}
	case 4:
		arg0, ok0 := canBeFloatImplicitly(args[0], argts[0])
		arg1, ok1 := canBeFloatImplicitly(args[1], argts[1])
		arg2, ok2 := canBeFloatImplicitly(args[2], argts[2])
		arg3, ok3 := canBeFloatImplicitly(args[3], argts[3])
		if ok0 && ok1 && ok2 && ok3 {
			args[0], args[1], args[2], args[3] = arg0, arg1, arg2, arg3
			return nil
		}
	default:
		return fmt.Errorf("invalid number of arguments for vec4")
	}

	var str []string
	for _, t := range argts {
		str = append(str, t.String())
	}
	return fmt.Errorf("invalid arguments for vec4: (%s)", strings.Join(str, ", "))
}

func checkArgsForMat2BuiltinFunc(args []shaderir.Expr, argts []shaderir.Type) error {
	if len(args) != len(argts) {
		return fmt.Errorf("the number of arguments and types doesn't match: %d vs %d", len(args), len(argts))
	}

	switch len(args) {
	case 1:
		if arg, ok := canBeFloatImplicitly(args[0], argts[0]); ok {
			args[0] = arg
			return nil
		}
		if argts[0].Main == shaderir.Mat2 {
			return nil
		}
	case 2:
		if argts[0].IsVector() && argts[0].VectorElementCount() == 2 && argts[1].IsVector() && argts[1].VectorElementCount() == 2 {
			return nil
		}
	case 4:
		ok := true
		for i := range argts {
			if _, canBeFloat := canBeFloatImplicitly(args[i], argts[i]); !canBeFloat {
				ok = false
				break
			}
		}
		if ok {
			for i := range argts {
				args[i], _ = canBeFloatImplicitly(args[i], argts[i])
			}
			return nil
		}
	default:
		return fmt.Errorf("invalid number of arguments for mat2")
	}

	var str []string
	for _, t := range argts {
		str = append(str, t.String())
	}
	return fmt.Errorf("invalid arguments for mat2: (%s)", strings.Join(str, ", "))
}

func checkArgsForMat3BuiltinFunc(args []shaderir.Expr, argts []shaderir.Type) error {
	if len(args) != len(argts) {
		return fmt.Errorf("the number of arguments and types doesn't match: %d vs %d", len(args), len(argts))
	}

	switch len(args) {
	case 1:
		if arg, ok := canBeFloatImplicitly(args[0], argts[0]); ok {
			args[0] = arg
			return nil
		}
		if argts[0].Main == shaderir.Mat3 {
			return nil
		}
	case 3:
		if argts[0].IsVector() && argts[0].VectorElementCount() == 3 &&
			argts[1].IsVector() && argts[1].VectorElementCount() == 3 &&
			argts[2].IsVector() && argts[2].VectorElementCount() == 3 {
			return nil
		}
	case 9:
		ok := true
		for i := range argts {
			if _, canBeFloat := canBeFloatImplicitly(args[i], argts[i]); !canBeFloat {
				ok = false
				break
			}
		}
		if ok {
			for i := range argts {
				args[i], _ = canBeFloatImplicitly(args[i], argts[i])
			}
			return nil
		}
	default:
		return fmt.Errorf("invalid number of arguments for mat3")
	}

	var str []string
	for _, t := range argts {
		str = append(str, t.String())
	}
	return fmt.Errorf("invalid arguments for mat3: (%s)", strings.Join(str, ", "))
}

func checkArgsForMat4BuiltinFunc(args []shaderir.Expr, argts []shaderir.Type) error {
	if len(args) != len(argts) {
		return fmt.Errorf("the number of arguments and types doesn't match: %d vs %d", len(args), len(argts))
	}

	switch len(args) {
	case 1:
		if arg, ok := canBeFloatImplicitly(args[0], argts[0]); ok {
			args[0] = arg
			return nil
		}
		if argts[0].Main == shaderir.Mat4 {
			return nil
		}
	case 4:
		if argts[0].IsVector() && argts[0].VectorElementCount() == 4 &&
			argts[1].IsVector() && argts[1].VectorElementCount() == 4 &&
			argts[2].IsVector() && argts[2].VectorElementCount() == 4 &&
			argts[3].IsVector() && argts[3].VectorElementCount() == 4 {
			return nil
		}
	case 16:
		ok := true
		for i := range argts {
			if _, canBeFloat := canBeFloatImplicitly(args[i], argts[i]); !canBeFloat {
				ok = false
				break
			}
		}
		if ok {
			for i := range argts {
				args[i], _ = canBeFloatImplicitly(args[i], argts[i])
			}
			return nil
		}
	default:
		return fmt.Errorf("invalid number of arguments for mat4")
	}

	var str []string
	for _, t := range argts {
		str = append(str, t.String())
	}
	return fmt.Errorf("invalid arguments for mat4: (%s)", strings.Join(str, ", "))
}
