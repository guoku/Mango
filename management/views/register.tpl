 <div class="container-fluid">
    <form class="form-horizontal well" method="POST" action="/register">
        <div class="control-group">
            <p class="text-error">{{.flash.error}}</p>
        </div>
        <div class="control-group">
            <label class="control-label" for="user_email">Email</label>
            <div class="controls">
                <input id="user_email" class="input-medium" value="{{.Invitation.Email}}" readonly="readonly">
            </div>
        </div>
        <div class="control-group">
            <label class="control-label" for="user_pass">Password</label>
            <div class="controls">
                <input id="user_pass" name="password" type="password" class="input-medium">
            </div>
        </div>
        <div class="control-group">
            <label class="control-label" for="user_name">Your Name</label>
            <div class="controls">
                <input id="user_name" name="name" type="text" class="input-medium">
            </div>
        </div>
        <div class="control-group">
            <label class="control-label" for="user_nick">Nickname</label>
            <div class="controls">
                <input id="user_nick" name="nickname" type="text" class="input-medium">
            </div>
        </div>
        <div class="control-group">
            <label class="control-label" for="user_mobile">Mobile</label>
            <div class="controls">
                <input id="user_mobile" name="mobile" type="text" class="input-medium">
            </div>
        </div>
        <div class="control-group">
            <label class="control-label" for="user_department">Department</label>
            <div class="controls">
                <select id="user_department" name="department">
                    <option>Engineering</option>
                    <option>Marketing</option>
                    <option>Product</option>
                    <option>Operation</option>
                    <option>Other</option>
                </select>
            </div>
        </div>
        <input type="hidden" name="token" value="{{.Invitation.Token}}">
        
        <div class="control-group">
            <div class ="controls">
            <button type="submit" class="btn">Submit</button>
            </div>
        </div>
    </form>
</div>
