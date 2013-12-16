{{ template "nav.tpl" .}}
<div class="container-fluid">
        {{template "dict_nav.tpl" .DictTab}}
    <div class="span10">
        <form class="form-search" method="GET" action="/dict_manage/brands/">
            <input type="text" class="input-medium search-query" name="q" value="{{ .SearchQuery }}">
            <button type="submit" class="btn">查找词语</button>
        </form>   
        <form id="addword" class="form-search" method="POST" action="/dict_manage/brands/add/">
            <input type="text" class="input-medium search-query" name="w">
            <button type="submit" class="btn">添加词语</button>
        </form>
        <div class="alert">
          <p id="addresult"></p>
        </div>
        <table class="table table-bordered table-striped">
            <th>Word</th>
            <th>Frequency</th>
            <th>Valided</th>
            <th>delete</th>
            {{ range .Words }}
                <tr 
                        {{if .Valid }}
                                class="success"
                              {{else }}}
                                class="warning" 
                              {{end}}
                >
                    <td>
                              {{ .Name }}
                    </td>
                    <td>{{ .Freq }}</td>
                    <td>
                        <form method="POST" class="narrow-form update_form" margin-bottom="0px" action="/dict_manage/brands/update/">
                        <input type="hidden" name="w" value="{{.Name}}">
                        {{ if .Valid }}
                        <input class="valided_value" type="hidden" name="valid" value="false">
                        <button type="submit" class="btn">取消确定为品牌</button>
                        {{ else }}
                        <input class="valided_value" type="hidden" name="valid" value="true">
                        <button type="submit" class="btn">确定为品牌吗?</button>
                        {{ end }}
                        </form>
                    </td>
                    <td>
                        <form method="POST" class="narrow-form delete_form" margin-bottom="0px" action="/dict_manage/brands/delete/">
                        <input type="hidden" name="w" value="{{.Name}}">
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
            dataType : 'json',
            success : function(data) {
                if (data.error) { 
                    alert("Error")
                    return
                }
                if (data.valid) {
                    f.children("button").text("取消确定为品牌")
                    f.children(".valided_value").attr("value", "false")
                } else {
                    f.children("button").text("确定为品牌吗?")
                    f.children(".valided_value").attr("value", "true")
                }
                var word = f.parent().parent()
                word.removeClass()
                if (data.deleted) {
                    word.addClass("error")
                } else if (data.valid) {
                    word.addClass("warning")
                } else {
                    word.addClass("success")
                }
            }
        });
        ev.preventDefault();
    });
    var dfrm = $(".delete_form")
    dfrm.submit(function (ev) {
        var f = $(this)
        $.ajax({
            type : f.attr('method'),
            url : f.attr('action'),
            data : f.serialize(),
            dataType : 'json',
            success : function(data) {
                if (data.error) { 
                    alert("Error")
                    return
                }
                var pf = $(".update_form")
                if (data.deleted) {
                    f.children("button").text("恢复")
                    f.children(".delete_value").attr("value", "false")
                    pf.children("button").text("确定为品牌吗?")
                    pf.children("button").disabled = true
                    pf.children(".delete_value").attr("value","true")
                } else {
                    f.children("button").text("删除")
                    f.children(".delete_value").attr("value", "true")
                } 
                var word = f.parent().parent()
                word.removeClass()
                if (data.deleted) {
                    word.addClass("error")
                } else if (data.blacklisted) {
                    word.addClass("warning")
                } else {
                    word.addClass("success")
                }
            }
        });
        ev.preventDefault();
    });

    var addForm = $("#addword")
    addForm.submit(function(ev) {
        var f = $(this)
        $.ajax({
            type : f.attr('method'),
            url : f.attr('action'),
            data : f.serialize(),
            success : function(data) {
                p = $("#addresult")
                if (data == "Success") {
                    p.html("<strong>添加成功</strong>")
                } else if (data == "Existed") {
                    p.html("<strong>该词已经存在</strong>")
                } else {
                    p.html("<strong>服务器错误</strong>" + data)
                }
            }
        });
        ev.preventDefault();
    });

</script>

{{ template "paginator.tpl" .}}
