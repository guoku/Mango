{{ template "nav.tpl" .}}
<div class="container-fluid">
    <div class="span10">
        <table class="table table-bordered table-striped">
            <th>Word</th>
            <th>Frequency</th>
            <th>Weight</th>
            <th>blacklist</th>
            <th>delete</th>
            {{ range .Words }}
                <tr>
                    <td>{{ .Word }}</td>
                    <td>{{ .Freq }}</td>
                    <td>{{ .Weight }}</td>
                    <td>
                        <form method="POST" class="update_form form-inline" action="/scheduler/dict_manage/update/">
                        <input type="hidden" name="w" value="{{.Word}}">
                        {{ if .Blacklisted }}
                        <input class="blacklist_value" type="hidden" name="blacklist" value="false">
                        <button type="submit" class="btn">取消黑名单</button>
                        {{ else }}
                        <input class="blacklist_value" type="hidden" name="blacklist" value="true">
                        <button type="submit" class="btn">加入黑名单</button>
                        {{ end }}
                        </form>
                    </td>
                    <td>
                        <form method="POST" class="delete_form form-inline" action="/scheduler/dict_manage/delete/">
                        <input type="hidden" name="w" value="{{.Word}}">
                        {{ if .Deleted }}
                        <input class="delete_value" type="hidden" name="delete" value="false">
                        <button type="submit" class="btn">恢复</button>
                        {{ else }}
                        <input class="delete_value" type="hidden" name="delete" value="true">
                        <button type="submit" class="btn">删除</button>
                        {{ end }}
                        </form>
                    </td>
                </tr>
            {{ end }}
        </table>
    </div>
</div>

<script>
    var frm = $(".update_form")
    frm.submit(function (ev) {
        var f = $(this)
        $.ajax({
            type : f.attr('method'),
            url : f.attr('action'),
            data : f.serialize(),
            success : function(data) {
                if (data == "0") {
                    f.children("button").text("取消黑名单")
                    f.children(".blacklist_value").attr("value", "false")
                } else if (data == "1") {
                    f.children("button").text("加入黑名单")
                    f.children(".blacklist_value").attr("value", "true")
                } else {
                    alert(data)
                }
            }
        });
        ev.preventDefault();
        return false
    });
    var dfrm = $(".delete_form")
    dfrm.submit(function (ev) {
        var f = $(this)
        $.ajax({
            type : f.attr('method'),
            url : f.attr('action'),
            data : f.serialize(),
            success : function(data) {
                if (data == "0") {
                    f.children("button").text("恢复")
                    f.children(".delete_value").attr("value", "false")
                } else if (data == "1") {
                    f.children("button").text("删除")
                    f.children(".delete_value").attr("value", "true")
                } else {
                    alert(data)
                }
            }
        });
        ev.preventDefault();
        return false
    });

</script>

{{ template "paginator.tpl" .}}
