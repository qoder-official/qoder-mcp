import 'package:get_it/get_it.dart';
import 'package:template_project_name/app/feature_templates/auth/mvc/controller/auth_controller.provider.dart';

final GetIt locator = GetIt.instance;

Future<void> setupDependencies() async {
  // Register the AuthController as a factory; each call makes a new instance.
  locator.registerFactory(() => AuthController());
}
