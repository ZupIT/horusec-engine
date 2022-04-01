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

function IfStatement(a, b) {
    if (a >= 10) {
        a = b * 2
    } else if(a <= 5) {
        a = b + a;
    } else {
        a = a + b
        const c = a * 10;
        console.log(c);
    }
    return a;
}

function TryStatement() {
    console.log('try entry')

    try {
        console.log('try body')
    }
    catch (e) {
        console.log(e)
        console.log('try catch')
    }
    finally {
        console.log('try finally')
    }

    console.log('try done')
}

function TryStatementWithoutFinally() {
    console.log('try entry')

    try {
        console.log('try body')
    }
    catch (e) {
        console.log(e)
        console.log('try catch')
    }

    console.log('try done')
}

function TryStatementWithoutCatch() {
    console.log('try entry')

    try {
        console.log('try body')
    }
    finally {
        console.log('try finally')
    }

    console.log('try done')
}

function WhileStatement() {
    let i = 0;
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
        i++;
    }
    console.log("finish")
}

function LabeledWhileStatement() {
    let x = 0;
    whileStmt :while ( x <= 5){
        console.log('test')
        if (i === 4 ){
            console.log("finish")
            break whileStmt
        }
        if (i === 2){
            console.log("two")
            continue whileStmt
        }
        x++;
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

function ForStatementIteratingOverList(data) {
    let sum = 0;
    for (let i =0; i < data.length; i++) {
        sum += i;
    }

    return sum
}

function ForStatementWithoutBinaryExpressionIncremet() {
    for (var a, b; c; d)
        e;
}

function ForStatementEndlessRecursion() {
    for (;;) {
        console.log("endless recursion");
    }
}

function ForStatementEmptyBody() {
    for (var i = 0
        ; i < l
        ; i++) {
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
}
