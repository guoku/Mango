{{ template "nav.tpl" . }}

<div class="container-fluid">
    <div class="span10">
        {{ if .Item.Title }}
        <h3><a href="{{.Item.DetailUrl}}">{{ .Item.Title }}</a></h3>
        {{ else }}
        <h3>Data Not Ready</h3>
        {{ end }}
        <div class="span8">
            <h3>API Data</h3>
            <table class="table table-bordered table-striped">
                <tr>
                    <td>NumIid</td>
                    <td><a href="http://item.taobao.com/item.html?id={{.Item.NumIid }}">{{.Item.NumIid}} </a> </td>
                </tr>
                {{ if .Item.Title }}
                <tr>
                    <td>Shop</td>
                    <td><a href="http://shop{{.Item.Sid}}.taobao.com">{{.Item.Nick}}</a></td>
                </tr>
                <tr>
                    <td>Cid</td>
                    <td>{{.Item.Cid}}</td>
                </tr>
                <tr>
                    <td>Price</td>
                    <td>{{.Item.Price}}</td>
                </tr>
                <tr>
                    <td>Location</td>
                    <td>{{.Item.Location}}</td>
                </tr>
                <tr>
                    <td>Props</td>
                    <td>{{.Item.Props}}</td>
                </tr>
                <tr>
                    <td>Images</td>
                    <td>
                        {{ range .Item.ItemImgs }}
                            <img src="{{.}}" width="100" height="100">
                        {{ end }}
                    </td>
                <tr>
                <tr>
                    <td>Desc</td>
                    <td>{{ str2html .Item.Desc}}</td>
                </tr>
                {{ else }}
                <tr> <td>Data Not Ready </td>  <td>Data Not Ready </td> </tr>
                {{ end }}

            </table>
        </div>
        <div class="span8">
            <h3>Crawled Data</h3>
            <table class="table table-bordered table-striped">
            </table>
        </div>
    </div>
</div>
