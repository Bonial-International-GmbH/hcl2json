package convert

import (
	"encoding/json"
	"sync"
)

// Bulk converts multiple files concurrently. The resulting byte slice contains
// a JSON object keyed by file path, e.g.:
//
//   {
//     "path/to/file.tf":{"resource":{"aws_route53_record":{"type": "CNAME", ...}}}},
//     "path/to/otherfile.tf":{"variable":{"foo":{"value": "bar"}}}
//   }
//
func Bulk(concurrency int, paths []string, options Options) ([]byte, error) {
	if len(paths) == 0 {
		return []byte(`{}`), nil
	}

	if concurrency < 0 {
		concurrency = 1
	} else if len(paths) < concurrency {
		concurrency = len(paths)
	}

	workCh := make(chan string)
	outCh := make(chan *convertResult, concurrency)
	errCh := make(chan error, concurrency)

	defer close(errCh)

	var wg sync.WaitGroup

	// Start workers.
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for path := range workCh {
				converted, err := File(path, options)
				if err != nil {
					errCh <- err
					return
				}

				outCh <- &convertResult{
					path:  path,
					bytes: converted,
				}
			}
		}()
	}

	// Assign work.
	go func() {
		defer close(outCh)

		for _, path := range paths {
			workCh <- path
		}

		close(workCh)

		wg.Wait()
	}()

	fileMap := make(map[string]json.RawMessage)

	// Collect results.
	for {
		select {
		case res, ok := <-outCh:
			if !ok {
				// outCh was closed, we are done.
				return json.Marshal(fileMap)
			}

			// Treat result bytes as raw JSON to prevent double encoding.
			fileMap[res.path] = json.RawMessage(res.bytes)
		case err := <-errCh:
			return nil, err
		}
	}
}

type convertResult struct {
	path  string
	bytes []byte
}
