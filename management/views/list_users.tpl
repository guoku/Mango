{{ template "nav.tpl"}}

<div class="container-fluid">
    <div class="page-header">
        <h2>User List</h2>
    </div>
    <table class="table table-striped table-bordered table-hover">
        <thead>
            <tr>
                <th>#</th>
                <th>Name</th>
                <th>Nickname</th>
                <th>Email</th>
                <th>Mobile</th>
                <th>Department</th>
            </tr>
        </thead>
        <tbody>
            {{ range .Users}}
            <tr>
                <td>{{.Id}}</td>
                <td>{{.Name}}</td>
                <td>{{.Nickname}}</td>
                <td>{{.Email}}</td>
                <td>{{.Additional.Mobile}}</td>
                <td>{{.Additional.Department}}</td>
            </tr>
            {{ end }}
        </tbody>
    </table>
</div>
