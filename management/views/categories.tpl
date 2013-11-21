{{ template "nav.tpl" .}}
<div class="container-fluid">
    <div class="span10">
        <a href="/commodity/category_manage/">类目管理</a>
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
                <li><a href="/commodity/category/?taobao_cid={{.ItemCat.Cid}}">{{.ItemCat.Name}}</a>(Cid:{{.ItemCat.Cid}})({{.ItemNum}})(精选:{{.SelectionNum}})<span class="divider">.</span></li>
                {{ end }}
            </ul>
        </div>
        
        {{ else }}
        {{ if .HasSubCats}}
        <div class="span9">
            <ul class="breadcrumb">
                {{ range .DirectSubCats }}
                <li><a href="/commodity/category/?taobao_cid={{.ItemCat.Cid}}">{{.ItemCat.Name}}</a>(Cid:{{.ItemCat.Cid}})({{.ItemNum}})(精选:{{.SelectionNum}})<span class="divider">.</span></li>
                {{ end }}
            </ul>
        </div>
        {{end}}
        <div class="span9">
            <table class="table table-bordered table-striped">
                <th>Select</th>
                <th>Taobao ID</th>
                <th>Image</th>
                <th>Title</th>
                <th>Link</th>
                <th>Price</th>
                <th>Score</th>
                {{ range .Items }}
                <tr>
                    <td>
                        {{ if  .ItemId  }}
                        <input type="checkbox" class="mul_taobao_id" value="{{.NumIid}}">
                        {{ else }}
                        Uploaded
                        {{ end }}
                    </td>
                    <td><a href="/scheduler/item_detail/taobao/?id={{.NumIid}}">{{.NumIid}}</a></td>
                    {{ if .Title }}
                        <td>
                        {{if .ItemImgs}}
                            <img src="{{index .ItemImgs 0}}" width="100" height="100"/>
                        {{end}}
                        </td>
                        <td>{{.Title}}</td>
                        <td><a href="{{.DetailUrl}}">Link</a></td>
                        <td>{{.Price}}</td>
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
            <button id="add_items">Add Items</button>
        </div>
        {{ end }}
    </div>
</div>
<script>
    $(document).ready(function () {
        $("#add_items").click(function() {
            var boxes = $(".mul_taobao_id")
            var ids = []
            boxes.each(function() {
                if (this.checked) {
                    ids.push($(this).attr("value"))
                }
            });
            //alert(ids.join(","))
            $.post("/commodity/add_online_items/",
                    {
                        taobao_ids:ids.join(","),
                        cid: "{{.Cid}}"
                    },
                    function(data, status) {
                        if (status == "success") {
                            window.location.href=window.location.href;
                        }
                        else {
                            alert(status)
                        }
                    });
        });
    });
</script>
{{ template "paginator.tpl" .}}

