{{ template "nav.tpl" .}}

<div class="container-fluid">
    <div class="span10">
        <form class="well form-search" method="GET" action="/commodity/category/">
            <input type="text" class="input-medium search-query" name="q">
            <button type="submit" class="btn">搜索类目</button>
        </form>   
        
        <div class="span9">
            <ul class="breadcrumb">
                <li><a href="/commodity/category/">All</a></li>
                {{ if eq .IsSearch false }} 
                {{ range .CatsPath }}
                <li><span class="divider">/</span><a href="/commodity/category/?taobao_cid={{.ItemCat.Cid}}">{{.ItemCat.Name}}</a></li>
                {{ end }}
                {{ end }}
            </ul>
        </div>
        {{ if .IsSearch }}
        <div class="span9">
            <ul class="breadcrumb">
                {{ range .SearchCats }}
                <li><a href="/commodity/category/?taobao_cid={{.ItemCat.Cid}}">{{.ItemCat.Name}}</a>(Cid:{{.ItemCat.Cid}})({{.ItemNum}})<span class="divider">.</span></li>
                {{ end }}
            </ul>
        </div>
        
        {{ else }}
        {{ if .HasSubCats}}
        <div class="span9">
            <ul class="breadcrumb">
                {{ range .DirectSubCats }}
                <li><a href="/commodity/category/?taobao_cid={{.ItemCat.Cid}}">{{.ItemCat.Name}}</a>(Cid:{{.ItemCat.Cid}})({{.ItemNum}}) <span class="divider">.</span></li>
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
        {{ end }}
    </div>
</div>
{{ template "paginator.tpl" .}}
