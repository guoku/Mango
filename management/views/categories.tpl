{{ template "nav.tpl" .}}

<div class="container-fluid">
    <div class="span10">
        <div class="span9">
            <ul class="breadcrumb">
                <li><a href="/commodity/category">All</a></li>
                {{ range .CatsPath }}
                <li><span class="divider">/</span><a href="/commodity/category/?taobao_cid={{.ItemCat.Cid}}">{{.ItemCat.Name}}</a></li>
                {{ end }}
            </ul>
        </div>
        {{ if .HasSubCats}}
        <div class="span9">
            <ul class="breadcrumb">
                {{ range .DirectSubCats }}
                <li><a href="/commodity/category/?taobao_cid={{.ItemCat.Cid}}">{{.ItemCat.Name}}({{.ItemNum}}) <span class="divider">.</span></a></li>
                {{ end }}
            </ul>
        </div>
        {{end}}
        <div class="span9">
            <table class="table table-bordered table-striped">
                <th>Taobao ID</th>
                <th>Image</th>
                <th>Title</th>
                <th>Link</th>
                <th>Price</th>
                <th>Score</th>
                {{ range .Items }}
                <tr>
                    <td><a href="/scheduler/item_detail/taobao/?id={{.NumIid}}">{{.NumIid}}</a></td>
                    {{ if .ApiDataReady }}
                        {{with index .ApiData.ItemImgs.ItemImgArray 0 }}
                            <td><img src="{{.Url}}" width="100" height="100"/></td>
                        {{end}}
                        <td>{{.ApiData.Title}}</td>
                        <td><a href="{{.ApiData.DetailUrl}}">Link</a></td>
                        <td>{{.ApiData.Price}}</td>
                        <td>{{.Score}}</td>
                    {{ else }}
                        <td>not ready</td>
                        <td>not ready</td>
                        <td>not ready</td>
                        <td>not ready</td>
                        <td>not ready</td>
                    {{ end }}
                </tr>
                {{ end }}
            </table>
        </div>
    </div>
</div>
{{ template "paginator.tpl" .}}
