## HTTP Request Tool
This tool allows you to make HTTP requests and print the address of the request along with the MD5 hash of the response. The tool is designed to perform the requests in parallel, allowing it to complete faster. It also includes a feature to limit the number of parallel requests to prevent exhausting local resources.

### Usage
Command Line Arguments
The tool accepts the following command line arguments:

-parallel: The maximum number of parallel requests (default is 10).

### Examples
To run the tool with default settings (10 parallel requests), simply run:

```
$ ./myhttp http://example.com https://google.com http://github.com
```

To run the tool with a maximum of 5 parallel requests and verbose output, run:
```
$ ./myhttp -parallel 5 -v http://example.com https://google.com http://github.com
```

### Building and Running the Tool
To build and run the tool, simply follow these steps:

* clone this repository
* inside this repository, build the tool using `make`
* run the tool with a list of URLs: `./myhttp http://example.com https://google.com http://github.com`

