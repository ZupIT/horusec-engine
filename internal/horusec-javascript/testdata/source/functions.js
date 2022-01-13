function f1(a, b) {}

const f2 = (a, b) => {}

function f3(a, b) {
	const c = a + b;
	console.log(c);
	const foo = doFoo(c);
	return foo;
}

function f4() {
	const sum = (a,b) => {
		return a + b;
	};
	const sub = function(a,b) {
		return a - b;
	}
	const result = doSomething(sum(10,20));
	console.log(` result: ${result}`);
}
