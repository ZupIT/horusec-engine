// Copyright 2022 ZUP IT SERVICOS EM TECNOLOGIA E INOVACAO SA
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

package pool

import (
	"time"

	"github.com/panjf2000/ants/v2"
)

const (
	// DefaultAntsPoolSize sets up the capacity of worker pool, 256 * 1024.
	DefaultAntsPoolSize = 10

	// ExpiryDuration is the interval time to clean up those expired workers.
	ExpiryDuration = 10 * time.Second
)

// Pool is the alias of ants.Pool.
type Pool = ants.Pool

// NewPool instantiates a new goroutine pool with poolSize argument or default pool size.
func NewPool(poolSize int) (*Pool, error) {
	return ants.NewPool(getDefaultOrInformedPoolSize(poolSize), ants.WithOptions(getOptions()))
}

// getDefaultOrInformedPoolSize returns informed pool size if greater than 0 or default pool size if 0 or lower
func getDefaultOrInformedPoolSize(poolSize int) int {
	if poolSize > 0 {
		return poolSize
	}

	return DefaultAntsPoolSize
}

// getOptions get ants goroutine pool options
func getOptions() ants.Options {
	return ants.Options{
		ExpiryDuration: ExpiryDuration,
	}
}
