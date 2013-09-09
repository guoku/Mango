{{ template "nav.tpl" }}
<div>
    <a href="/add_pass" role="button" class="btn btn-large btn-primary">添加密码</a>
    <br><br> 
    <!-- Modal -->
    <table class="table table-striped table-bordered table-condensed">
        <thead>
            <tr>
                <th> Name </th>
                <th> Account </th>
                <th> Password </th>
                <th> Desc </th>
                <th> Operation </th>
            </tr>
        </thead>
        <tbody>
            {{ range .User.PasswordPermissions }}
                <tr>
                    <td>{{ .Password.Name }}</td>
                    <td>{{ .Password.Account }}</td>
                    <td>{{ .Password.Password }}</td>
                    <td>{{ .Password.Desc }}</td>
                    <td>
                        <a href="/edit_pass" class="btn btn-large btn-primary">Edit</a>
                    </td>
                <tr>
            {{end}}
        </tbody>
    </table>
</div>
