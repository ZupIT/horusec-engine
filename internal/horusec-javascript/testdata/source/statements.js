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
}
