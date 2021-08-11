var appPath = 'app/';
var buildPath = 'build/';

module.exports = function (grunt) {
  'use strict';

  // Grunt init
  grunt.initConfig({
  	// Clean build directory
  	clean: {
	    options: {
		    force: true
	     },
	    all: [buildPath]
  	},
    browserify: {
      options: {
        transform: ['reactify']
	    },
      dist: {
        src: [appPath + 'gateway/src/**/*.js', appPath + 'gateway/src/**/*.jsx'],
        dest: buildPath + 'gateway/src/app.js'
	    }
    },
    // Copy files over to build directory
    copy: {
      templates: {
        files: [{
          expand: true,
          flatten: false,
          cwd: appPath,
          src: 'templates/*.{html,js}',
          dest: buildPath
        }]
      },
      gateway: {
        files: [{
          expand: true,
          flatten: false,
          cwd: appPath + 'gateway/html',
          src: '*.html',
          dest:  buildPath + 'gateway'
        },
        {
  		    expand: true,
  		    flatten: false,
  		    cwd: appPath + 'gateway',
          src: 'css/*.css',
  		    dest: buildPath + 'gateway'
        },
        {
  		    expand: true,
  		    flatten: false,
  		    cwd: appPath + 'gateway',
  		    src: 'assets/**/*.*',
  		    dest: buildPath + 'gateway'
        }]
      },
      liquid: {
        files: [{
  		    expand: true,
  		    flatten: false,
  		    cwd: appPath,
  		    src: 'liquid/*.html',
  		    dest: buildPath
        }]
      },
      admin: {
        files: [{
          expand: true,
          flatten: false,
          cwd: appPath,
          src: 'admin/*.{html,css,js}',
          dest: buildPath
        },
        {
          expand: true,
          flatten: false,
          cwd: appPath,
          src: 'admin/assets/**/*.*',
          dest: buildPath
        }]
      }
    },
    watch: {
      templates: {
        files: [appPath + 'templates/*.{html,js}'],
        tasks: ['clean', 'copy']
      },
      scripts: {
        files: [appPath + 'gateway/src/**/*.js', appPath + 'gateway/src/**/*.jsx'],
        tasks: ['clean', 'copy', 'browserify']
      },
      css: {
        files: [appPath + 'gateway/css/*.css'],
        tasks: ['clean', 'copy', 'browserify']
      },
      admin: {
        files: [appPath + 'admin/**/*.{html,css,js}'],
        tasks: ['clean', 'copy']
      }
    }
  });

  require('matchdep').filterDev('grunt-*').forEach(grunt.loadNpmTasks);
  
  // register grunt tasks
  grunt.registerTask('default', ['clean', 'copy', 'browserify']);
};
