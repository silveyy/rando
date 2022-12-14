using Android.App;
using Android.Runtime;

namespace maui_test;

[Application]
public class MainApplication : MauiApplication
{
	public MainApplication(IntPtr handle, JniHandleOwnership ownership)
		: base(handle, ownership)
	{

		// how do I do somework here.

		
	}

	protected override MauiApp CreateMauiApp() => MauiProgram.CreateMauiApp();
}
