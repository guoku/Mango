{{ template "nav.tpl" . }}

<div class="container-fluid">
    <div class="span10">
        <h3><a href="{{.Item.ApiData.DetailUrl}}">{{ .Item.ApiData.Title }}</a></h3>
        <div class="span8">
            <h3>API Data</h3>
            <table class="table table-bordered table-striped">
                <tr>
                    <td>NumIid</td>
                    <td>{{.Item.ApiData.NumIid }} </td>
                </tr>
                <tr>
                    <td>Shop</td>
                    <td><a href="http://shop{{.Sid}}.taobao.com">{{.Item.ApiData.Nick}}</a></td>
                </tr>
                <tr>
                    <td>Type</td>
                    <td>{{.Item.ApiData.Type}}</td>
                </tr>
                <tr>
                    <td>Cid</td>
                    <td>{{.Item.ApiData.Cid}}</td>
                </tr>
                <tr>
                    <td>Price</td>
                    <td>{{.Item.ApiData.Price}}</td>
                </tr>
                <tr>
                    <td>GlobalStockType</td>
                    <td>{{.Item.ApiData.GlobalStockType}}</td>
                </tr>
                <tr>
                    <td>Location</td>
                    <td>{{.Item.ApiData.Location}}</td>
                </tr>
                <tr>
                    <td>Props</td>
                    <td>{{.Item.ApiData.PropsName}}</td>
                </tr>
                <tr>
                    <td>Images</td>
                    <td>
                        {{ range .Item.ApiData.ItemImgs.ItemImgArray }}
                            <img src="{{.Url}}" width="100" height="100">
                        {{ end }}
                    </td>
                <tr>
                <tr>
                    <td>Desc</td>
                    <td>{{ str2html .Item.ApiData.Desc}}</td>
                </tr>

            </table>
        </div>
        <div class="span8">
            <h3>Crawled Data</h3>
            <table class="table table-bordered table-striped">
            </table>
        </div>
    </div>
</div>
