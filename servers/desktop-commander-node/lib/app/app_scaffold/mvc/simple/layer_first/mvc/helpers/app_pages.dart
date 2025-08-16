import 'package:flutter/material.dart';
import 'package:template_project_name/app/app_scaffold/mvc/simple/layer_first/mvc/constants/route_constants.dart';
import 'package:template_project_name/app/app_scaffold/mvc/simple/layer_first/mvc/view/login_screen.dart';
import 'package:template_project_name/app/app_scaffold/mvc/simple/layer_first/mvc/view/home_screen.dart';

class AppPages {
  static const String initial = RouteConstants.loginRoute;

  static final Map<String, WidgetBuilder> routes = {
    RouteConstants.loginRoute: (context) => const LoginScreen(),
    RouteConstants.homeRoute: (context) => const HomeScreen(),
  };

  static Route<dynamic>? onGenerateRoute(RouteSettings settings) {
    final WidgetBuilder? builder = routes[settings.name];
    if (builder != null) {
      return MaterialPageRoute(builder: builder, settings: settings);
    }
    // Handle unknown routes gracefully.
    return MaterialPageRoute(
      builder: (_) => Scaffold(
        body: Center(
          child: Text('No route defined for ${settings.name}'),
        ),
      ),
    );
  }
}
