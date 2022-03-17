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

// Function documentation
// Suspendisse potenti. Integer porttitor elit non urna viverra pretium. Mauris aliquam condimentum tempus. Nullam mi 
// massa, porttitor nec tincidunt nec, posuere dapibus libero. Sed id tortor purus.
//
// #nosec
function f1() {}

/* 
    * #nosec
*/
function f2() {}

function f3() {
    // #nosec // This syntax is supported.
    console.log("f3-1 ignored")

    console.log("f3-2") // #nosec // This syntax is not supported.

    console.log("f3-3" /* #nosec */)  // This syntax is not supported.
}

class Foo {
    bar() {
    // #nosec 
    console.log("Foo.bar");
    }

    // #nosec
    baz() {}
}


// #nosec
class Bar {}


// #nosec
const a = 20

