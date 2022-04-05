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

console.log('testing');

const express = require('express')

const app = express()

app.get('/', (req, res) => {
    console.log(req, res)
});

app.set('/', (req, res) => {
    console.log(req, res)
});


function incrementExpr() {
    let a = 0;
    a++
    a--
    return a
}

function subscriptExpr() {
    const values = ['a', 'b', 'c'];

    let i = 0

    console.log(values[i])
    console.log(values[1])
}
