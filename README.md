XSLgen
======

Compiled language to generate XSLs templates

Usage
------

Build the project and run it :

```
go build && ./XSLgen --file "/path/to/your/xslgen/file.xslg"
```

Example
------

``` python
@?xml [version: "1.0", encoding: "UTF-8"]
@stylesheet [ xsl: "http://xsl", tif: "http://tif", dc: "http://dc", dcterms: "http://dcterms" ] {
	@output [ method: "xml", omit-xml-declaration: "no", indent: "yes", encoding: "UTF-8" ] {
		# Test
		&dc.DublinCore {
			&dc.title : "My title"
			&dc.author : "Axel Anceau"
		}
	}
}
```

Output :
```
<?xml version="1.0" encoding="UTF-8"/>
<xsl:stylesheet xsl="http://xsl" tif="http://tif" dc="http://dc" dcterms="http://dcterms">
	<xsl:output method="xml" omit-xml-declaration="no" indent="yes" encoding="UTF-8">
		<dc:DublinCore>
			<dc:title>My title</dc:title>
			<dc:author>Axel Anceau</dc:author>
		</dc:DublinCore>
	</xsl:output>
</xsl:stylesheet>
```

Enjoy :)