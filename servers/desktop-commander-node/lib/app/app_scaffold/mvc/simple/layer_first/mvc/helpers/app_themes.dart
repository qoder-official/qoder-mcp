import 'package:flutter/material.dart';
import 'package:template_project_name/app/common/constants/color_constants.dart';

/// Defines light and dark themes for the layer-first MVC sample.
class AppThemes {
  static final ThemeData lightTheme = ThemeData(
    brightness: Brightness.light,
    primaryColor: ColorConstants.primary,
    scaffoldBackgroundColor: ColorConstants.lightScaffoldBackground,
    appBarTheme: const AppBarTheme(
      backgroundColor: ColorConstants.lightAppBarBackground,
      iconTheme: IconThemeData(color: ColorConstants.lightAppBarIconColor),
      titleTextStyle: TextStyle(
        color: ColorConstants.lightAppBarTitleColor,
        fontSize: 20,
        fontWeight: FontWeight.bold,
      ),
    ),
  );

  static final ThemeData darkTheme = ThemeData(
    brightness: Brightness.dark,
    primaryColor: ColorConstants.primary,
    scaffoldBackgroundColor: ColorConstants.darkScaffoldBackground,
    appBarTheme: const AppBarTheme(
      backgroundColor: ColorConstants.darkAppBarBackground,
      iconTheme: IconThemeData(color: ColorConstants.darkAppBarIconColor),
      titleTextStyle: TextStyle(
        color: ColorConstants.darkAppBarTitleColor,
        fontSize: 20,
        fontWeight: FontWeight.bold,
      ),
    ),
  );
}
