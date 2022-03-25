/**
 * Copyright 2022 ZUP IT SERVICOS EM TECNOLOGIA E INOVACAO SA
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import fs from 'fs';

function f1(a, b) { return f2(a+b) * 5 }

const f2 = (a) => { return a + 10; }

function f3(a, b) {
    let a = 20;
	const c = a + b;
	console.log(c);
	const res = f2(c);
	return res;
}

function f4() {
	const sum = (a,b) => {
		return a + b;
	};
	const sub = function(a,b) {
		return a - b;
	}
	const result = doSomething(sum(10,20));
	console.log(`result: ${result}`);
}

function f5(path) {
    fs.readFile(path, (err, data) => {
        console.log(data);
    });
}

function f6() {
    const value = f1(10, 20) + f2(30);
    return value * 10 / 2;
}

const a = "x"

function f7(b) {
	const c = "y"
	const d = c
	const e = d
	const f = a
	const g = f
	const h = b
}

function f8(a) {
	const b = a
	const c = b
}

function f9(a) {
    a.b().c(10)
    a.b.c()

    const a = a.b(10).c.d(20)

    const b = a.b.c.d.e

    a.b = c()
}

function f10(){
	var x = Math.random()
	console.log(x)
	var y = "z"
	console.log(y)
}

function f11() {
	let a = ['a', 'b', 'c']
	let b = new ExClass()
	let c = new ExClass('a', 'b', 'c', f2())
	let d = { 'k': 'v', 'k2': 'v2' }
	let e = [{ 'k': 'v' }, { 'k': 'v' }]
}

const i = [1,2,3]

function f12() {
	console.log(i)
}

function f13() {
    let a = new A();
    a.foo()
}
