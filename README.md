gowiki
========

basic web server expanded from the tutorial:
- https://golang.org/doc/articles/wiki/

## Usage

`bin/wiki <port=num> <log=level>`

## Options

- port, default: 8080
- log level, default: info

## Tasks

- Implement inter-page linking by converting instances of [PageName] to
<a href="/view/PageName">PageName</a>. (hint: you could use regexp.ReplaceAllFunc to do this)
- Spruce up the page templates by making them valid HTML and adding some CSS rules.
- Add a home button which links to root page
