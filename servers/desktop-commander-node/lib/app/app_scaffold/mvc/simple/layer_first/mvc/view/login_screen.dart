import 'package:flutter/material.dart';
import 'package:template_project_name/app/app_scaffold/mvc/simple/layer_first/mvc/constants/route_constants.dart';

class LoginScreen extends StatelessWidget {
  const LoginScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Login')),
      body: Center(
        child: ElevatedButton(
          onPressed: () {
            Navigator.pushNamed(context, RouteConstants.homeRoute);
          },
          child: const Text('Proceed to Home'),
        ),
      ),
    );
  }
}
