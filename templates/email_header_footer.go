package templates

func EmailHeader() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8" />
    <title>Email</title>
    <style>
        body {
            background-color: #301B68;
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 0;
        }
        .header {
            background-color: rgba(48, 27, 104, 0.80);
            padding: 20px;
            text-align: center;
        }
        .header img {
            height: 40px;
        }
        .content {
            color: #fff;
            margin: 20px;
        }
        .footer {
            color: #ffff;
            background-color: rgba(48, 27, 104, 0.80);
            padding: 10px;
            text-align: center;
        }
        .footer a {
            margin: 0 15px;
            text-decoration: none;
            color: #fff;
        }
        .footer a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="header">
        <img src="https://tarangpadia.github.io/magdaTestImg/SurrealXPLogo.png" alt="Company Logo" />
    </div>
    <div class="content">
`
}

func EmailFooter() string {
	return `
    </div>
    <div class="footer">
        <a href="#">About Us</a> |
        <a href="#">Privacy Policy</a> |
        <a href="#">Contact</a>
    </div>
</body>
</html>
`
}
