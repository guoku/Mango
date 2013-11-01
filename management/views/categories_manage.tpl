{{ template "nav.tpl" .}}

<div class="container-fluid">
    <div class="span10">
        <form method="GET" action="/commodity/category_manage/">
            <input type="hidden" name="update" value="true">
            <button type="submit" class="btn">Update</button>
        </form>
        <div class="span9">
            <table class="table table-bordered table-striped">
                <th> Category ID </th>
                <th> Icon </th>
                <th> Title </th>
                <th> Matched Taobao Categories </th>
                <th> Operations </th>
                {{ range .GuokuCats }}
                <tr>
                    <td>{{.CategoryId}}</td>
                    <td><img src="{{.IconSmall}}"></td>
                    <td>{{.Title}}</td>
                    <td>
                        {{ range .MatchedTaobaoCats }}
                            <a href="/commodity/category/?taobao_cid={{.ItemCat.Cid}}">{{.ItemCat.Name}}</a>
                            <br>
                        {{ end }}
                    </td>
                    <td>
                        <form method="POST" class="form-inline" action="/commodity/category_manage/add_taobao_category/">
                            <input type="hidden" name="guoku_cid" value="{{.CategoryId}}">
                            <input class="input-small" type="text" name="taobao_cid" width="80">
                            <button type="submit" class="btn">Add</button>
                        </form>
                    </td>
                </tr>
                {{ end }}
            </table>
        </div>
    </div>
</div>
