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

function TryStatement() {
    try {
        const value = 'test'
        console.log(value)
    } catch (err) {
        console.error(err)
    } finally {
        let sum = 1 + 1
        console.log(sum)
    }
}

function WhileStatement() {
    whileStmt :while ( i <= 5){
        console.log('test')
        if (i === 4 ){
            console.log("finish")
            break whileStmt
        }
        if (i === 2){
            console.log("two")
            continue whileStmt
        }
    }
    while ( i <= 5){
        console.log('test')
        if (i === 2){
            console.log("two")
            continue
        }
        if (i === 4 ){
            console.log("finish")
            break
        }
    }
}

function SwitchStatement() {
    let fruits = 'Oranges'

    switch (fruits) {
        case 'Oranges':
            console.log('Oranges')
            break
        case 'Mangoes':
            console.log('Mangoes')
            break
        case 'Papayas':
            console.log('Papayas')
            break
        default:
            console.log('No fruits')
    }
}

function ForStatement() {
    for (let i = 0; i < 9; i++) {
        console.log(i);
    }
}

function ForInStatement() {
    const values = ['a', 'b', 'c'];

    for (let value in values) {
        console.log(value)
    }
}


function ExportStatement() {
    export let test1, test2, test3

    export let test4 = 'test4', test5 = 'test5', test6;

    export function testFunc1() {
        console.log('test')
    }

    export class TestExportClass {
        constructor() {
            console.log('test')
        }

    }

    export {test7, test8, test9};

    export {test10 as alias10, test11 as alias11, test12};

    export default test13;

    export default function () {
        console.log('test')
    }

    export default function testFunc2() {
        console.log('test')
    }

    export {name14 as default, nome15};

    export * from 'test1.js';

    export * as name16 from 'test2.js';

    export {name17, name18, name19} from 'test3.js';

    export {name20 as alias20, name21 as alias21, name22} from 'test4.js';

    export {default, name23} from 'test5.js';
}
