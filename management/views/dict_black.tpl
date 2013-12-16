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
                        <form method="POST" class="update_form form-inline" action="/dict_manage/update/">
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
                        <form method="POST" class="delete_form form-inline" action="/dict_manage/delete/">
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
