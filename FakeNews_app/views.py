from django.shortcuts import render
from django.contrib.auth.mixins import LoginRequiredMixin, UserPassesTestMixin
from .models import Post
from ctypes import *
import ctypes

from django.views.generic import (
    ListView,
    DetailView,
    CreateView,
    UpdateView,
    DeleteView
)


def home(request):
    context = {
        'posts': Post.objects.all()
    }
    return render(request, 'FakeNews_app/home.html', context)

def about(request):
    return render(request, 'FakeNews_app/about.html',{'title':'About page'})


class PostListView(ListView):
    model = Post
    template_name = 'FakeNews_app/home.html'  # <app>/<model>_<viewtype>.html
    context_object_name = 'posts'
    ordering = ['-date_posted']


class PostDetailView(DetailView):
    model = Post


class PostCreateView(LoginRequiredMixin, CreateView):
    model = Post
    fields = ['title', 'content']

    def form_valid(self, form):
        form.instance.author = self.request.user
        return super().form_valid(form)


class PostUpdateView(LoginRequiredMixin, UserPassesTestMixin, UpdateView):
    model = Post
    fields = ['title', 'content']

    def form_valid(self, form):
        form.instance.author = self.request.user
        return super().form_valid(form)

    def test_func(self):
        post = self.get_object()
        if self.request.user == post.author:
            return True
        return False


class PostDeleteView(LoginRequiredMixin, UserPassesTestMixin, DeleteView):
    model = Post
    success_url = '/'

    def test_func(self):
        post = self.get_object()
        if self.request.user == post.author:
            return True
        return False



def SignView(request, pk):
    allPosts=Post.objects.all()

    mesaage = Post.objects.get(id=pk).content

    aggrated_signature = cosi(str(mesaage))
    
    context = {
        'posts': allPosts,
        'pk':pk,
        'signature':aggrated_signature.decode(),
    }


    return render(request, 'FakeNews_app/signCosi.html', context)


def cosi(message):

    lib = cdll.LoadLibrary("./Gocode/src/github.com/MHDRateb/cosi_test/cosiTest.so")
    
    # define class GoString to map:
    # C type struct { const char *p; GoInt n; }
    class GoString(Structure):
        _fields_ = [("p", c_char_p), ("n", c_longlong)]  

    lib.startSign.argtypes = [GoString]
    lib.startSign.restype = c_longlong  
    msg = GoString( str.encode(message), len(message))
    signaggr = lib.startSign(msg)
    # print ('returned value',ctypes.string_at(signaggr))
    return ctypes.string_at(signaggr)
    
    
    
    
    