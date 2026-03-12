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
            background-color: #fff;
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 0;
        }
        .header {
            background-color: #fff;
            padding: 20px;
            text-align: center;
        }
        .header img {
            height: 40px;
        }
        .content {
            color: #1a1a1a;
            margin: 20px;
        }
        .footer {
            color: #1a1a1a;
            background-color: #fff;
            padding: 10px;
            text-align: center;
        }
        .footer a {
            margin: 0 15px;
            text-decoration: none;
            color: #1a1a1a;
        }
        .footer a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="header">
        <img src="https://tarangpadia.github.io/magdaTestImg/numerra_logo.png" alt="Company Logo" />
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
