Coconut automatically organizes photos in a hardrive without much intervention from the user. It is non-destructive and deterministic.

### Installation

Make sure you have a working Go environment.  Go version 1.11+ is supported. [See the install instructions for Go](http://golang.org/doc/install.html).

Go Modules are strongly recommended when using this package. [See the go blog guide on using Go Modules](https://blog.golang.org/using-go-modules).

```
$ GO111MODULE=on go get github.com/sergi/coconut
```

### How it works

Given one or more source folders and a destination folder, Coconut goes through all the image files in the source folders and organizes them in the destination folder using their EXIF metadata, de-duplicating them during the process.Coconut **never** deletes or modifies files in the source folders.

Coconut organizes photos using a folder hierarchy. The default hierarchy is as follows:

```
Year
└──Year-Month
	 └──Geographical Place
	 		└──original_filename.jpeg
```

An example folder hierarchy could look like this, realistically:

```
2018
└──September
   └───South Lake Tahoe-US
   │   ├──DSC_1735.jpeg
   │   └──DSC_2187.raw
   └───Sonoma-US
       ├──DSC_2395.jpeg
       └──DSC_0934.cr3
```

The folder hierachy can be changed in `config.yml`. Instructions on how to change it are in "Changing the path template".



### Usage

```
coconut /source_folder1 /source_folder2 --destination /destination_folder
```



### Changing the path template

TBD

###License

MPL 2.0